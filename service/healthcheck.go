package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/therealak12/api-health-check/config"
	"net/http"
	"sync"
	"time"

	"github.com/therealak12/api-health-check/repository"
)

const (
	healthcheckDefaultTimeout = 5
	webhookDefaultTimeout     = 5
)

type cancelMap struct {
	sync.Mutex
	funcs map[int]context.CancelFunc
}

func newCancelMap() *cancelMap {
	return &cancelMap{
		funcs: make(map[int]context.CancelFunc),
	}
}

func (c *cancelMap) Get(key int) (value context.CancelFunc, ok bool) {
	c.Lock()
	result, ok := c.funcs[key]
	c.Unlock()
	return result, ok
}

func (c *cancelMap) Set(key int, value context.CancelFunc) {
	c.Lock()
	c.funcs[key] = value
	c.Unlock()
}

func (c *cancelMap) Delete(key int) {
	c.Lock()
	delete(c.funcs, key)
	c.Unlock()
}

var (
	healthchecks = newCancelMap()
)

type HealthcheckService interface {
	StartHealthCheck(healthcheckID int) error
	StoptHealthCheck(healthcheckID int) error
}

type healthcheckService struct {
	healthcheckRepo      repository.HealthcheckRepo
	healthcheckEventRepo repository.HealthcheckEventRepo
	webhookConfig        config.Webhook
}

var _ HealthcheckService = &healthcheckService{}

func NewHealthcheckService(healthcheckRepo repository.HealthcheckRepo,
	healthcheckEventRepo repository.HealthcheckEventRepo,
	webhookConfig config.Webhook) HealthcheckService {
	return &healthcheckService{
		healthcheckRepo:      healthcheckRepo,
		healthcheckEventRepo: healthcheckEventRepo,
		webhookConfig:        webhookConfig,
	}
}

func (hs *healthcheckService) StartHealthCheck(healthcheckID int) error {
	healthcheck, err := hs.healthcheckRepo.FindOne(healthcheckID)
	if err == repository.ErrRecordNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get healthcheck")
	}

	ticker := time.NewTicker(time.Duration(healthcheck.IntervalSeconds) * time.Second)

	httpClient := http.Client{Timeout: healthcheckDefaultTimeout * time.Second}

	req, err := http.NewRequest(
		healthcheck.HttpMethod,
		healthcheck.Url,
		bytes.NewBuffer([]byte(healthcheck.Body)),
	)
	if err != nil {
		return err
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(healthcheck.HeadersJson), &headers); err != nil {
		return err
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	checkAPIHealth := func() {
		resp, err := httpClient.Do(req)
		var status string
		if err != nil {
			logrus.Warnf("failed to make healthcheck request, err: %s", err)
			status = fmt.Sprintf("healthcheck failed, err: %s", err)
		} else {
			status = resp.Status
			//goland:noinspection GoUnhandledErrorResult
			defer resp.Body.Close()
		}
		healthcheckEvent := repository.HealthcheckEvent{
			HealthcheckID: healthcheckID,
			Status:        status,
			CreatedAt:     time.Time{},
		}
		lastHealthcheckEvent, err := hs.healthcheckEventRepo.FindLast()
		if err != nil && err != repository.ErrRecordNotFound {
			logrus.Errorf("failed to get last healthcheck event, err: %s", err)
			return
		}
		if err == repository.ErrRecordNotFound {
			if err := hs.healthcheckEventRepo.Create(&healthcheckEvent); err != nil {
				logrus.Errorf("failed to create healthcheck event, err: %s", err)
				return
			}
		} else {
			if differs := hs.compareHealthcheckEvents(&lastHealthcheckEvent, &healthcheckEvent); differs {
				if err := hs.sendHealthStatusAlert(&lastHealthcheckEvent, &healthcheckEvent); err != nil {
					logrus.Errorf("failed to send healthcheck alert, err: %s", err)
				}
			}
			if err := hs.healthcheckEventRepo.Create(&healthcheckEvent); err != nil {
				logrus.Errorf("failed to create healthcheck event, err: %s", err)
				return
			}
		}
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	healthchecks.Set(healthcheck.ID, cancelFunc)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logrus.Debugf("stopping health check, id: %d", healthcheckID)
				break
			case <-ticker.C:
				logrus.Debugf("checking api health, id: %d", healthcheckID)
				checkAPIHealth()
			}
		}
	}()

	return nil
}

func (hs *healthcheckService) compareHealthcheckEvents(lastHealthcheckEvent, healthcheckEvent *repository.HealthcheckEvent) bool {
	return lastHealthcheckEvent.Status != healthcheckEvent.Status
}

func (hs *healthcheckService) sendHealthStatusAlert(lastHealthcheckEvent, healthcheckEvent *repository.HealthcheckEvent) error {
	httpClient := http.Client{Timeout: webhookDefaultTimeout * time.Second}

	req, err := http.NewRequest(
		"POST",
		hs.webhookConfig.Url,
		bytes.NewBuffer([]byte(fmt.Sprintf(
			`{"%s":"health status changed, was %s and is %s"}`,
			hs.webhookConfig.MessageFieldName,
			lastHealthcheckEvent.Status,
			healthcheckEvent.Status,
		))),
	)
	if err != nil {
		return err
	}

	_, err = httpClient.Do(req)
	return err
}

func (hs *healthcheckService) StoptHealthCheck(healthcheckID int) error {
	_, err := hs.healthcheckRepo.FindOne(healthcheckID)
	if err == repository.ErrRecordNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get healthcheck")
	}
	cancelFunc, ok := healthchecks.Get(healthcheckID)
	if !ok {
		return errors.New(fmt.Sprintf("the health check %d is not started yet", healthcheckID))
	}

	cancelFunc()
	healthchecks.Delete(healthcheckID)

	return nil
}

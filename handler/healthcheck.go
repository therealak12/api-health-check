package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/therealak12/api-health-check/repository"
	"github.com/therealak12/api-health-check/request"
	"github.com/therealak12/api-health-check/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// HealthcheckHandler handles operations defined for healthcheck.
type HealthcheckHandler struct {
	HealthcheckRepo    repository.HealthcheckRepo
	HealthcheckService service.HealthcheckService
}

func NewHealthcheckHandler(healthcheckRepo repository.HealthcheckRepo,
	healthcheckService service.HealthcheckService) HealthcheckHandler {
	return HealthcheckHandler{
		HealthcheckRepo:    healthcheckRepo,
		HealthcheckService: healthcheckService,
	}
}

func (h HealthcheckHandler) Register(c echo.Context) error {
	req := &request.CreateHealthcheck{}

	if err := c.Bind(req); err != nil {
		logrus.Errorf("create healthcheck: bind failed: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bind request failed: %s", err))
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bad request: %s", err.Error()))
	}

	headersJson, err := json.Marshal(req.Headers)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bad request: %s", err.Error()))
	}

	healthcheck := &repository.Healthcheck{
		IntervalSeconds: req.IntervalSeconds,
		Url:             req.Url,
		HttpMethod:      req.HttpMethod,
		HeadersJson:     string(headersJson),
		Body:            req.Body,
	}

	if err := h.HealthcheckRepo.Save(healthcheck); err != nil {
		logrus.Errorf("failed to create healthcheck: %s", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create healthcheck")
	}

	return c.JSON(http.StatusCreated, healthcheck)
}

func (h HealthcheckHandler) Start(c echo.Context) error {
	req := &request.ToggleHealthcheck{}

	if err := c.Bind(req); err != nil {
		logrus.Errorf("start healthcheck: bind failed: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bind request failed: %s", err))
	}

	if err := h.HealthcheckService.StartHealthCheck(req.ID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "")
}

func (h HealthcheckHandler) Stop(c echo.Context) error {
	req := &request.ToggleHealthcheck{}

	if err := c.Bind(req); err != nil {
		logrus.Errorf("start healthcheck: bind failed: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bind request failed: %s", err))
	}

	if err := h.HealthcheckService.StoptHealthCheck(req.ID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "")
}

func (h HealthcheckHandler) List(c echo.Context) error {
	healthchecks, err := h.HealthcheckRepo.FindAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list healthchecks")
	}

	return c.JSON(http.StatusOK, healthchecks)
}

func (h HealthcheckHandler) Delete(c echo.Context) error {
	req := &request.DeleteHealthcheck{}

	if err := c.Bind(req); err != nil {
		logrus.Errorf("delete healthcheck: bind failed: %s", err.Error())

		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bind request failed: %s", err))
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bad request: %s", err.Error()))
	}

	if err := h.HealthcheckRepo.Delete(req.ID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "healthcheck id not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete healthcheck")
	}

	return c.NoContent(http.StatusOK)
}

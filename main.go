package main

import (
	"context"
	"github.com/therealak12/api-health-check/handler"
	"github.com/therealak12/api-health-check/repository"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/therealak12/api-health-check/config"
	"github.com/therealak12/api-health-check/database"
	"github.com/therealak12/api-health-check/request"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	server := echo.New()

	server.Validator = request.NewValidator()

	server.Use(middleware.Recover())
	server.Use(middleware.CORS())
	server.Use(middleware.RemoveTrailingSlash())

	cfg := config.New()
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(cfg.Logger.Level)

	db, err := database.NewPostgresInstance(cfg.Database)
	if err != nil {
		logrus.Fatalf("failed to connect to database: %s", err.Error())
	}
	healthCheckRepo := repository.SQLHealthcheckRepo{
		DB: db,
	}

	healthCheckHandler := handler.HealthcheckHandler{HealthcheckRepo: healthCheckRepo}

	server.GET("/healthchecks", healthCheckHandler.List)
	server.POST("/healthchecks", healthCheckHandler.Register)
	server.GET("/healthchecks/:id/start", healthCheckHandler.Start)
	server.GET("/healthchecks/:id/stop", healthCheckHandler.Stop)
	server.DELETE("/healthchecks/:id", healthCheckHandler.Delete)

	go func() {
		err := server.Start(":8080")
		if err != nil {
			logrus.Fatalf("failed to start server: %s", err.Error())
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	s := <-sig
	logrus.Infof("got signal %s, shutting down", s)

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown gracefully: %s", err.Error())
	}
}

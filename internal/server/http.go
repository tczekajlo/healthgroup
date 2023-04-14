package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/tczekajlo/healthgroup/internal/config"
	handler "github.com/tczekajlo/healthgroup/internal/handlers"
	flogger "github.com/tczekajlo/healthgroup/internal/middleware/logger"
	"github.com/tczekajlo/healthgroup/internal/version"
	"go.uber.org/zap"
)

func NewHTTP(config *config.Config, logger *zap.Logger) error {
	logger.Info("healthgroup", zap.String("version", version.Version))

	addr := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		IdleTimeout:           config.Server.IdleTimeout,
		AppName:               fmt.Sprintf("healthgroup/%s", version.Version),
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(flogger.New(flogger.Config{
		Logger: logger,
	}))

	// Routes
	app.Get("/health/kubernetes/:namespace/:service", handler.HealthKubernetes(config, logger))
	app.Get("/health/consul/:namespace/:service", handler.HealthConsul(config, logger))
	app.Get("/health/consul/:service", handler.HealthConsul(config, logger))

	logger.Info("Listen", zap.String("addr", addr))

	go func() {
		if err := app.Listen(addr); err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	logger.Info("Gracefully shutting down...")
	if err := app.Shutdown(); err != nil {
		return err
	}

	logger.Info("healthgroup was successful shutdown")

	return nil
}

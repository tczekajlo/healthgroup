package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tczekajlo/healthgroup/internal/config"
	"github.com/tczekajlo/healthgroup/internal/discovery"
	"github.com/tczekajlo/healthgroup/internal/healthcheck"
	"go.uber.org/zap"
)

// HealthConsul is a function to run health check for a Consul service along with extra checks defined in the configuration file.
// @Summary Run health checks
// @Description Run health checks
// @Produce json
// @Param service path string true "Consul service"
// @Param namespace path string false "Consul namespace"
// @Success 200 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /health/consul/{service} [get]
// @Router /health/consul/{namespace}/{service} [get]
func HealthConsul(config *config.Config, logger *zap.Logger) fiber.Handler {
	h := &healthcheck.HealthCheck{
		Logger:    logger,
		Config:    config,
		Discovery: discovery.Consul,
	}

	d, err := discovery.New(&discovery.Discovery{
		Logger: logger,
		Config: config,
		Source: discovery.Consul,
	})
	if err != nil {
		return func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseHTTP{
				Success: false,
				Message: err.Error(),
			})
		}
	}
	defer d.Close()

	return func(c *fiber.Ctx) error {
		return Health(c, h, d)
	}
}

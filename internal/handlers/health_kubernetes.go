package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tczekajlo/healthgroup/internal/config"
	"github.com/tczekajlo/healthgroup/internal/discovery"
	"github.com/tczekajlo/healthgroup/internal/healthcheck"
	"go.uber.org/zap"
)

// HealthKubernetes is a function to run health check for a Kubernetes service along with extra checks defined in the configuration file.
// @Summary Run health checks
// @Description Run health checks
// @Produce json
// @Param namespace path string true "Kubernetes namespace"
// @Param service path string true "Kubernetes service"
// @Success 200 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /health/kubernetes/{namespace}/{service} [get]
func HealthKubernetes(config *config.Config, logger *zap.Logger) fiber.Handler {
	h := &healthcheck.HealthCheck{
		Logger: logger,
		Config: config,
	}

	d, err := discovery.New(&discovery.Discovery{
		Logger: logger,
		Config: config,
		Source: discovery.Kubernetes,
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

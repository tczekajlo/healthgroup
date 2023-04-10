package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tczekajlo/healthgroup/internal/discovery"
	"github.com/tczekajlo/healthgroup/internal/healthcheck"
)

func Health(c *fiber.Ctx, healthCheck *healthcheck.HealthCheck, discovery discovery.Adapter) error {
	exist, err := discovery.IsServiceExists(c)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseHTTP{
			Success: false,
			Message: err.Error(),
		})
	}

	if !exist {
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseHTTP{
			Success: false,
			Message: "Service doesn't exist",
		})
	}

	healthy, err := discovery.IsServiceHealthy(c)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseHTTP{
			Success: false,
			Message: err.Error(),
		})
	}

	if !healthy {
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseHTTP{
			Success: false,
			Message: "Service is not healthy",
		})
	}

	if err := healthCheck.Run(c); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseHTTP{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ResponseHTTP{
		Success: true,
		Message: "all health checks passed",
	})
}

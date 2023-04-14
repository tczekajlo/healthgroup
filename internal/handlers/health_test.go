package handlers

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/stretchr/testify/assert"
	"github.com/tczekajlo/healthgroup/internal/config"
	"github.com/tczekajlo/healthgroup/internal/log"
)

func TestRoute(t *testing.T) {
	logger, _ := log.NewAtLevel("ERROR")

	config := config.New(
		config.WithFile("../../test/testdata/config.yaml"),
		config.WithLogger(logger),
		config.WithFlags(&config.Flags{}),
	)
	errSetDefault := config.SetDefault()
	errReadConfig := config.ReadConfig()

	assert.Empty(t, errSetDefault)
	assert.Empty(t, errReadConfig)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(requestid.New())

	app.Get("/health/kubernetes/:namespace/:service", HealthKubernetes(config, logger))
	app.Get("/health/consul/:namespace/:service", HealthConsul(config, logger))
	app.Get("/health/consul/:service", HealthConsul(config, logger))

	table := []struct {
		desc         string
		path         string
		expectedCode int
	}{
		{
			desc:         "health check for Consul service",
			path:         "/health/consul/testservice",
			expectedCode: fiber.StatusServiceUnavailable,
		},
		{
			desc:         "health check for Consul service with namespace",
			path:         "/health/consul/testns/testservice",
			expectedCode: fiber.StatusServiceUnavailable,
		},
		{
			desc:         "health check for Kubernetes",
			path:         "/health/kubernetes/testns/testservice",
			expectedCode: fiber.StatusServiceUnavailable,
		},
	}

	for _, item := range table {
		t.Run(item.desc, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost%s", item.path), nil)
			resp, _ := app.Test(req)

			assert.Equal(t, item.expectedCode, resp.StatusCode)
		})
	}
}

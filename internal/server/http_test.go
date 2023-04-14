//go:build integration
// +build integration

package server

import (
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/stretchr/testify/assert"
	"github.com/tczekajlo/healthgroup/internal/config"
	"github.com/tczekajlo/healthgroup/internal/handlers"
	"github.com/tczekajlo/healthgroup/internal/log"
	"k8s.io/client-go/util/homedir"
)

func TestServer(t *testing.T) {
	logger, _ := log.NewAtLevel("ERROR")

	c := config.New(
		config.WithFile("../../test/testdata/config.yaml"),
		config.WithLogger(logger),
		config.WithFlags(&config.Flags{
			Kubeconfig: filepath.Join(homedir.HomeDir(), ".kube", "config"),
			InCluster:  false,
		}),
	)
	errSetDefault := c.SetDefault()
	errReadConfig := c.ReadConfig()

	assert.Empty(t, errSetDefault)
	assert.Empty(t, errReadConfig)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(requestid.New())

	app.Get("/health/kubernetes/:namespace/:service", handlers.HealthKubernetes(c, logger))
	app.Get("/health/consul/:namespace/:service", handlers.HealthConsul(c, logger))
	app.Get("/health/consul/:service", handlers.HealthConsul(c, logger))

	table := []struct {
		desc       string
		expected   bool
		path       string
		statusCode int
	}{
		{
			desc:       "Consul service unavailable",
			expected:   false,
			path:       "/health/consul/redis",
			statusCode: fiber.StatusServiceUnavailable,
		},
		{
			desc:       "Consul not found",
			expected:   false,
			path:       "/health/consul/testservice",
			statusCode: fiber.StatusNotFound,
		},
		{
			desc:       "Consul service unavailable namespace",
			expected:   false,
			path:       "/health/consul/ns/testservice",
			statusCode: fiber.StatusServiceUnavailable,
		},
		{
			desc:       "Consul service healthy",
			expected:   false,
			path:       "/health/consul/consul",
			statusCode: fiber.StatusOK,
		},
		{
			desc:       "Kubernetes not found",
			expected:   false,
			path:       "/health/kubernetes/test/test",
			statusCode: fiber.StatusNotFound,
		},
		{
			desc:       "Kubernetes service healthy",
			expected:   false,
			path:       "/health/kubernetes/default/kubernetes",
			statusCode: fiber.StatusOK,
		},
	}

	for _, item := range table {
		t.Run(item.desc, func(t *testing.T) {

			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost%s", item.path), nil)
			resp, err := app.Test(req)
			assert.Equal(t, item.statusCode, resp.StatusCode)
			assert.Empty(t, err)
		})
	}
}

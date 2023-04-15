package healthcheck

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/stretchr/testify/assert"
	"github.com/tczekajlo/healthgroup/internal/config"
	"github.com/tczekajlo/healthgroup/internal/discovery"
	"github.com/tczekajlo/healthgroup/internal/log"
)

func TestBuildURL(t *testing.T) {
	t.Parallel()

	table := []struct {
		desc     string
		check    config.HTTPHealthCheck
		expected string
	}{
		{
			desc: "HTTP",
			check: config.HTTPHealthCheck{
				Type: "http",
				Host: "example.com",
			},
			expected: "http://example.com",
		},
		{
			desc: "HTTP with a port",
			check: config.HTTPHealthCheck{
				Type: "http",
				Host: "example.com",
				Port: 8080,
			},
			expected: "http://example.com:8080",
		},
		{
			desc: "HTTP with a port and a path",
			check: config.HTTPHealthCheck{
				Type:        "http",
				Host:        "example.com",
				Port:        8080,
				RequestPath: "/test",
			},
			expected: "http://example.com:8080/test",
		},
		{
			desc: "HTTPS with a port and a path",
			check: config.HTTPHealthCheck{
				Type:        "https",
				Host:        "example.com",
				Port:        8080,
				RequestPath: "/test",
			},
			expected: "https://example.com:8080/test",
		},
		{
			desc: "HTTP2",
			check: config.HTTPHealthCheck{
				Type:        "http2",
				Host:        "example.com",
				Port:        8080,
				RequestPath: "/test",
			},
			expected: "https://example.com:8080/test",
		},
		{
			desc: "Unknown type",
			check: config.HTTPHealthCheck{
				Type: "tcp",
				Host: "example.com",
				Port: 8080,
			},
			expected: "",
		},
	}

	for _, item := range table {
		t.Run(item.desc, func(t *testing.T) {
			url, _ := buildURL(item.check)
			assert.Equal(t, item.expected, url)
		})
	}
}

func TestShouldSkip(t *testing.T) {
	t.Parallel()

	logger, _ := log.NewAtLevel("ERROR")

	c := config.New(
		config.WithFile("../../test/testdata/config.yaml"),
		config.WithLogger(logger),
		config.WithFlags(&config.Flags{}),
	)
	errSetDefault := c.SetDefault()
	errReadConfig := c.ReadConfig()

	assert.Empty(t, errSetDefault)
	assert.Empty(t, errReadConfig)

	check := HealthCheck{
		Logger:    logger,
		Config:    c,
		Discovery: discovery.Kubernetes,
	}

	table := []struct {
		desc        string
		healthCheck config.HTTPHealthCheck
		expected    bool
		path        string
		route       string
	}{
		{
			desc: "don't skip",
			healthCheck: config.HTTPHealthCheck{
				Type: "http",
				Host: "example.com",
			},
			expected: false,
			path:     "/health/consul/testservice",
			route:    "/health/consul/:service",
		},
		{
			desc: "match service",
			healthCheck: config.HTTPHealthCheck{
				Type:    "http",
				Host:    "example.com",
				Service: "testservice",
			},
			expected: false,
			path:     "/health/consul/testservice",
			route:    "/health/consul/:service",
		},
		{
			desc: "skip service",
			healthCheck: config.HTTPHealthCheck{
				Type:    "http",
				Host:    "example.com",
				Service: "testservice",
			},
			expected: true,
			path:     "/health/consul/testservice2",
			route:    "/health/consul/:service",
		},
		{
			desc: "kubernetes - skip service",
			healthCheck: config.HTTPHealthCheck{
				Type:    "http",
				Host:    "example.com",
				Service: "testservice",
			},
			expected: true,
			path:     "/health/kubernetes/ns/testservice2",
			route:    "/health/kubernetes/:namespace/:service",
		},
		{
			desc: "kubernetes - skip service and namespace",
			healthCheck: config.HTTPHealthCheck{
				Type:      "http",
				Host:      "example.com",
				Service:   "testservice",
				Namespace: "test",
			},
			expected: true,
			path:     "/health/kubernetes/ns/testservice2",
			route:    "/health/kubernetes/:namespace/:service",
		},
		{
			desc: "kubernetes - skip discovery",
			healthCheck: config.HTTPHealthCheck{
				Type:      "http",
				Host:      "example.com",
				Discovery: discovery.Consul,
			},
			expected: true,
			path:     "/health/kubernetes/ns/testservice2",
			route:    "/health/kubernetes/:namespace/:service",
		},
		{
			desc: "kubernetes - don't skip discovery",
			healthCheck: config.HTTPHealthCheck{
				Type:      "http",
				Host:      "example.com",
				Discovery: discovery.Kubernetes,
			},
			expected: false,
			path:     "/health/kubernetes/ns/testservice2",
			route:    "/health/kubernetes/:namespace/:service",
		},
		{
			desc: "kubernetes - match service and namespace",
			healthCheck: config.HTTPHealthCheck{
				Type:      "http",
				Host:      "example.com",
				Service:   "testservice2",
				Namespace: "ns",
			},
			expected: false,
			path:     "/health/kubernetes/ns/testservice2",
			route:    "/health/kubernetes/:namespace/:service",
		},
		{
			desc: "kubernetes - match namespace",
			healthCheck: config.HTTPHealthCheck{
				Type:      "http",
				Host:      "example.com",
				Namespace: "ns",
			},
			expected: false,
			path:     "/health/kubernetes/ns/testservice2",
			route:    "/health/kubernetes/:namespace/:service",
		},
	}

	for _, item := range table {
		t.Run(item.desc, func(t *testing.T) {
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})
			app.Use(requestid.New())
			app.Get(item.route, httpTestHandler(t, check, item.healthCheck, item.expected))

			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost%s", item.path), nil)
			_, err := app.Test(req)
			assert.Empty(t, err)
		})
	}
}

func httpTestHandler(t *testing.T, check HealthCheck, health config.HTTPHealthCheck, expected bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ex := check.shouldSkip(c, health)
		assert.Equal(t, expected, ex)
		return nil
	}
}

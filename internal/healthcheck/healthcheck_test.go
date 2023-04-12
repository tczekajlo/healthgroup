package healthcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tczekajlo/healthgroup/internal/config"
)

func TestBuildURL(t *testing.T) {
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

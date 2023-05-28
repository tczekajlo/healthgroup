//go:build integration
// +build integration

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tczekajlo/healthgroup/internal/log"
	"k8s.io/client-go/util/homedir"
)

func TestReadConfigMap(t *testing.T) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// To be sure that the env variables aren't set
	os.Unsetenv("HG_SERVER_ADDRESS")
	os.Unsetenv("HG_CONCURRENCY")
	os.Unsetenv("HG_SERVER_PORT")

	os.Setenv("HG_CONSUL_ENABLED", "true")

	logger, _ := log.NewAtLevel("ERROR")

	config := New(
		WithLogger(logger),
		WithFlags(&Flags{
			ConfigMap:  "default/healthgroup",
			Kubeconfig: kubeconfig,
		}),
	)
	errDefault := config.SetDefault()

	err := config.ReadConfig()

	assert.Equal(t, "0.0.0.0", config.Server.Address)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, 5, config.Concurrency)
	assert.Equal(t, HTTPHealthCheck{
		Type:        "http",
		Host:        "127.0.0.1",
		RequestPath: "/test",
		Port:        8500,
		Namespace:   "test_namespace",
		Timeout:     2 * time.Second,
		Service:     "test",
	}, config.HTTPHealthCheck[0])
	assert.Equal(t, false, config.Consul.Enabled)

	assert.Nil(t, errDefault, "error should be nil")
	assert.Nil(t, err, "error should be nil")
}

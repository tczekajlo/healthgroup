package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tczekajlo/healthgroup/internal/log"
)

func TestSetFromEnv(t *testing.T) {
	viper.SetEnvPrefix("healthgroup")
	viper.AutomaticEnv()

	config := New()

	os.Setenv("HEALTHGROUP_SERVER_ADDRESS", "testhost")
	os.Setenv("HEALTHGROUP_CONCURRENCY", "1")
	os.Setenv("HEALTHGROUP_SERVER_PORT", "123")
	os.Setenv("HEALTHGROUP_KUBERNETES_ENABLED", "true")
	os.Setenv("HEALTHGROUP_CONSUL_ENABLED", "true")
	os.Setenv("HEALTHGROUP_CONSUL_ADDRESS", "consul.host")
	os.Setenv("HEALTHGROUP_CONSUL_PORT", "8501")
	os.Setenv("HEALTHGROUP_CONSUL_SCHEME", "https")
	os.Setenv("HEALTHGROUP_CONSUL_INSECURE_SKIP_VERIFY", "true")
	os.Setenv("HEALTHGROUP_CONSUL_TOKEN", "testtoken")
	os.Setenv("HEALTHGROUP_CONSUL_CA_FILE", "/cafile")
	os.Setenv("HEALTHGROUP_CONSUL_CERT_FILE", "/certfile")
	os.Setenv("HEALTHGROUP_CONSUL_KEY_FILE", "/keyfile")
	os.Setenv("HEALTHGROUP_CONSUL_TIMEOUT", "10s")

	err := config.SetFromEnv()

	assert.Equal(t, "testhost", config.Server.Address, "HEALTHGROUP_SERVER_ADDRESS - should be equal")
	assert.Equal(t, 1, config.Concurrency, "HEALTHGROUP_CONCURRENCY - should be equal")
	assert.Equal(t, 123, config.Server.Port, "HEALTHGROUP_SERVER_PORT - should be equal")
	assert.Equal(t, true, config.Kubernetes.Enabled, "HEALTHGROUP_KUBERNETES_ENABLED - should be equal")
	assert.Equal(t, true, config.Consul.Enabled, "HEALTHGROUP_CONSUL_ENABLED - should be equal")
	assert.Equal(t, "consul.host", config.Consul.Address, "HEALTHGROUP_CONSUL_ADDRESS - should be equal")
	assert.Equal(t, 8501, config.Consul.Port, "HEALTHGROUP_CONSUL_PORT - should be equal")
	assert.Equal(t, "https", config.Consul.Scheme, "HEALTHGROUP_CONSUL_SCHEME - should be equal")
	assert.Equal(t, true, config.Consul.InsecureSkipVerify, "HEALTHGROUP_CONSUL_INSECURE_SKIP_VERIFY - should be equal")
	assert.Equal(t, "testtoken", config.Consul.Token, "HEALTHGROUP_CONSUL_TOKEN - should be equal")
	assert.Equal(t, "/cafile", config.Consul.CAFile, "HEALTHGROUP_CONSUL_TOKEN - should be equal")
	assert.Equal(t, "/certfile", config.Consul.CertFile, "HEALTHGROUP_CONSUL_CERT_FILE - should be equal")
	assert.Equal(t, "/keyfile", config.Consul.KeyFile, "HEALTHGROUP_CONSUL_KEY_FILE - should be equal")
	assert.Equal(t, time.Second*10, config.Consul.Timeout, "HEALTHGROUP_CONSUL_TIMEOUT - should be equal")

	assert.Nil(t, err, "error should be nil")
}

func TestSetDefault(t *testing.T) {
	t.Parallel()

	config := New()
	err := config.SetDefault()

	assert.Equal(t, "0.0.0.0", config.Server.Address)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, time.Second*5, config.Server.IdleTimeout)
	assert.Equal(t, 5, config.Concurrency)
	assert.Equal(t, true, config.Kubernetes.Enabled)
	assert.Equal(t, false, config.Consul.Enabled)
	assert.Equal(t, "127.0.0.1", config.Consul.Address)
	assert.Equal(t, "http", config.Consul.Scheme)
	assert.Equal(t, 8500, config.Consul.Port)
	assert.Empty(t, config.Consul.Token)
	assert.Equal(t, time.Second*2, config.Consul.Timeout)
	assert.Equal(t, false, config.Consul.InsecureSkipVerify)
	assert.Empty(t, config.Consul.CAFile)
	assert.Empty(t, config.Consul.CertFile)
	assert.Empty(t, config.Consul.KeyFile)

	assert.Nil(t, err, "error should be nil")
}

func TestReadConfig(t *testing.T) {
	// To be sure that the env variables aren't set
	os.Unsetenv("HEALTHGROUP_SERVER_ADDRESS")
	os.Unsetenv("HEALTHGROUP_CONCURRENCY")
	os.Unsetenv("HEALTHGROUP_SERVER_PORT")

	os.Setenv("HEALTHGROUP_CONSUL_ENABLED", "true")

	logger, _ := log.NewAtLevel("ERROR")

	config := New(
		WithFile("../../test/testdata/config.yaml"),
		WithLogger(logger),
		WithFlags(&Flags{}),
	)
	errDefault := config.SetDefault()

	err := config.ReadConfig()

	assert.Equal(t, "localhost", config.Server.Address)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, 5, config.Concurrency)
	assert.Equal(t, HTTPHealthCheck{
		Type:       "https",
		Host:       "google.com",
		TimeoutSec: 50,
		Service:    "test",
	}, config.HTTPHealthCheck[0])
	assert.Equal(t, true, config.Consul.Enabled)

	assert.Nil(t, errDefault, "error should be nil")
	assert.Nil(t, err, "error should be nil")
}

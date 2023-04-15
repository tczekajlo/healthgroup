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
	viper.SetEnvPrefix("hg")
	viper.AutomaticEnv()

	logger, _ := log.NewAtLevel("ERROR")
	config := New(WithLogger(logger))

	os.Setenv("HG_SERVER_ADDRESS", "testhost")
	os.Setenv("HG_CONCURRENCY", "1")
	os.Setenv("HG_SERVER_PORT", "123")
	os.Setenv("HG_KUBERNETES_ENABLED", "true")
	os.Setenv("HG_CONSUL_ENABLED", "true")
	os.Setenv("HG_CONSUL_ADDRESS", "consul.host:8500")
	os.Setenv("HG_CONSUL_SCHEME", "https")
	os.Setenv("HG_CONSUL_INSECURE_SKIP_VERIFY", "true")
	os.Setenv("HG_CONSUL_TOKEN", "testtoken")
	os.Setenv("HG_CONSUL_CA_FILE", "/cafile")
	os.Setenv("HG_CONSUL_CERT_FILE", "/certfile")
	os.Setenv("HG_CONSUL_KEY_FILE", "/keyfile")
	os.Setenv("HG_CONSUL_TIMEOUT", "10s")

	err := config.SetFromEnv()

	assert.Equal(t, "testhost", config.Server.Address, "HG_SERVER_ADDRESS - should be equal")
	assert.Equal(t, 1, config.Concurrency, "HG_CONCURRENCY - should be equal")
	assert.Equal(t, 123, config.Server.Port, "HG_SERVER_PORT - should be equal")
	assert.Equal(t, true, config.Kubernetes.Enabled, "HG_KUBERNETES_ENABLED - should be equal")
	assert.Equal(t, true, config.Consul.Enabled, "HG_CONSUL_ENABLED - should be equal")
	assert.Equal(t, "consul.host:8500", config.Consul.Address, "HG_CONSUL_ADDRESS - should be equal")
	assert.Equal(t, "https", config.Consul.Scheme, "HG_CONSUL_SCHEME - should be equal")
	assert.Equal(t, true, config.Consul.InsecureSkipVerify, "HG_CONSUL_INSECURE_SKIP_VERIFY - should be equal")
	assert.Equal(t, "testtoken", config.Consul.Token, "HG_CONSUL_TOKEN - should be equal")
	assert.Equal(t, "/cafile", config.Consul.CAFile, "HG_CONSUL_TOKEN - should be equal")
	assert.Equal(t, "/certfile", config.Consul.CertFile, "HG_CONSUL_CERT_FILE - should be equal")
	assert.Equal(t, "/keyfile", config.Consul.KeyFile, "HG_CONSUL_KEY_FILE - should be equal")
	assert.Equal(t, time.Second*10, config.Consul.Timeout, "HG_CONSUL_TIMEOUT - should be equal")

	assert.Nil(t, err, "error should be nil")
}

func TestSetDefault(t *testing.T) {
	t.Parallel()

	logger, _ := log.NewAtLevel("ERROR")
	config := New(WithLogger(logger))
	err := config.SetDefault()

	assert.Equal(t, "0.0.0.0", config.Server.Address)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, time.Second*5, config.Server.IdleTimeout)
	assert.Equal(t, 5, config.Concurrency)
	assert.Equal(t, true, config.Kubernetes.Enabled)
	assert.Equal(t, false, config.Consul.Enabled)
	assert.Equal(t, "127.0.0.1:8500", config.Consul.Address)
	assert.Equal(t, "http", config.Consul.Scheme)
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
	os.Unsetenv("HG_SERVER_ADDRESS")
	os.Unsetenv("HG_CONCURRENCY")
	os.Unsetenv("HG_SERVER_PORT")

	os.Setenv("HG_CONSUL_ENABLED", "true")

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
		Type:    "https",
		Host:    "google.com",
		Timeout: 5 * time.Second,
		Service: "test",
	}, config.HTTPHealthCheck[0])
	assert.Equal(t, true, config.Consul.Enabled)

	assert.Nil(t, errDefault, "error should be nil")
	assert.Nil(t, err, "error should be nil")
}

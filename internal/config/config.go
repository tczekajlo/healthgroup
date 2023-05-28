package config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func New(opts ...Option) *Config {
	config := &Config{}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// ReadConfigFromFile reads the configuration from a Kubernetes ConfigMap.
func (c *Config) ReadConfigFromConfigMap() error {
	var (
		config *rest.Config
		err    error
	)

	flags := c.Flags()

	if flags.InCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", c.flags.Kubeconfig)
		if err != nil {
			return err
		}
	}
	// creates the clientset
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfigAndClient(config, httpClient)
	if err != nil {
		return err
	}

	cfg := strings.Split(c.flags.ConfigMap, "/")

	if len(cfg) != 2 { //nolint:gomnd
		return xerrors.Errorf("ConfigMap doesn't exists: %s", cfg)
	}

	ns := cfg[0]
	name := cfg[1]

	c.logger.Info("Reading configuration from ConfigMap",
		zap.String("namespace", ns),
		zap.String("name", name),
	)
	cm, err := clientset.CoreV1().ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	r := bytes.NewBufferString(cm.Data["config.yaml"])
	viper.SetConfigType("yaml")

	return viper.ReadConfig(r)
}

// ReadConfigFromFile reads the configuration from a file.
func (c *Config) ReadConfigFromFile() {
	if c.file != "" {
		// Use config file from the flag.
		viper.SetConfigFile(c.file)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".healthgroup" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".healtgroup")
		viper.SetConfigType("yaml")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		c.logger.Info("Using config", zap.String("file", viper.ConfigFileUsed()))
	}
}

func (c *Config) ReadConfig() error {
	if err := c.SetDefault(); err != nil {
		return err
	}

	if c.flags.ConfigMap != "" {
		if err := c.ReadConfigFromConfigMap(); err != nil {
			return err
		}
	} else {
		c.ReadConfigFromFile()
	}

	err := viper.Unmarshal(&c)
	if err != nil {
		panic("configuration file: unable to decode into struct")
	}

	if err := c.SetFromEnv(); err != nil {
		return err
	}

	c.logger.Info("Configuration", zap.Any("config", c))

	return nil
}

func (c *Config) SetDefault() error {
	c.logger.Debug("setting default values")

	c.Server.Address = "0.0.0.0"
	c.Server.Port = 8080
	c.Server.IdleTimeout = time.Second * 5 //nolint:gomnd
	c.Concurrency = 5
	c.Kubernetes.Enabled = true
	c.Consul.Enabled = false
	c.Consul.Address = "127.0.0.1:8500"
	c.Consul.Scheme = "http"
	c.Consul.InsecureSkipVerify = false
	c.Consul.Timeout = time.Second * 2 //nolint:gomnd

	return nil
}

func (c *Config) SetFromEnv() error { //nolint:cyclop
	c.logger.Debug("reading environment variables and setting values")

	if v := viper.GetString("server_address"); v != "" {
		c.Server.Address = v
	}

	if v := viper.GetInt("server_port"); v != 0 {
		c.Server.Port = v
	}

	if v := viper.GetInt("concurrency"); v != 0 {
		c.Concurrency = v
	}

	_, ok := os.LookupEnv("HG_KUBERNETES_ENABLED")
	if v := viper.GetBool("kubernetes_enabled"); ok {
		c.Kubernetes.Enabled = v
	}

	_, ok = os.LookupEnv("HG_CONSUL_ENABLED")
	if v := viper.GetBool("consul_enabled"); ok {
		c.Consul.Enabled = v
	}

	if v := viper.GetString("consul_address"); v != "" {
		c.Consul.Address = v
	}

	if v := viper.GetString("consul_scheme"); v != "" {
		c.Consul.Scheme = v
	}

	if v := viper.GetString("consul_token"); v != "" {
		c.Consul.Token = v
	}

	if v := viper.GetString("consul_ca_file"); v != "" {
		c.Consul.CAFile = v
	}

	if v := viper.GetString("consul_cert_file"); v != "" {
		c.Consul.CertFile = v
	}

	if v := viper.GetString("consul_key_file"); v != "" {
		c.Consul.KeyFile = v
	}

	if v := viper.GetDuration("consul_timeout"); v != 0 {
		c.Consul.Timeout = v
	}

	_, ok = os.LookupEnv("HG_CONSUL_INSECURE_SKIP_VERIFY")
	if v := viper.GetBool("consul_insecure_skip_verify"); ok {
		c.Consul.InsecureSkipVerify = v
	}

	return nil
}

func (c *Config) Flags() *Flags {
	return c.flags
}

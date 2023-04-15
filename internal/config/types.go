package config

import (
	"time"

	"go.uber.org/zap"
)

type Option func(c *Config)

type Flags struct {
	Kubeconfig string
	ConfigFile string
	InCluster  bool
}

type Config struct {
	logger          *zap.Logger
	file            string
	flags           *Flags
	Server          Server
	HTTPHealthCheck []HTTPHealthCheck
	Concurrency     int
	Kubernetes      Kubernetes
	Consul          Consul
}

type Server struct {
	Address     string
	Port        int
	IdleTimeout time.Duration
}

type HTTPHealthCheck struct {
	Timeout            time.Duration
	Type               string
	RequestPath        string
	Port               int
	Host               string
	Service            string
	Namespace          string
	InsecureSkipVerify bool
	Discovery          string
}

type Kubernetes struct {
	Enabled bool
}

type Consul struct {
	Enabled            bool
	Address            string
	Scheme             string
	CAFile             string
	CertFile           string
	KeyFile            string
	InsecureSkipVerify bool
	Token              string
	Timeout            time.Duration
}

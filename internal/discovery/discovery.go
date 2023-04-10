package discovery

import (
	"github.com/tczekajlo/healthgroup/internal/discovery/consul"
	"github.com/tczekajlo/healthgroup/internal/discovery/k8s"
)

const (
	Kubernetes = "kubernetes"
	Consul     = "consul"
)

func New(discovery *Discovery) (Adapter, error) {
	switch discovery.Source {
	case Kubernetes:
		return k8s.New(&k8s.Client{
			Logger: discovery.Logger,
			Config: discovery.Config,
		})
	case Consul:
		return consul.New(&consul.Client{
			Logger: discovery.Logger,
			Config: discovery.Config,
		})
	default:
		return nil, nil
	}
}

package consul

import (
	"github.com/gofiber/fiber/v2"
	capi "github.com/hashicorp/consul/api"
	"github.com/tczekajlo/healthgroup/internal/config"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

type Client struct {
	Logger *zap.Logger
	Config *config.Config

	client       *capi.Client
	consulConfig *capi.Config
}

func New(c *Client) (*Client, error) {
	if !c.Config.Consul.Enabled {
		return nil, xerrors.New("Consul client is disabled. You can enabled it in the configuration file")
	}

	c.consulConfig = capi.DefaultConfig()

	switch c.Config.Consul.Scheme {
	case "https":
		c.consulConfig.TLSConfig.Address = c.Config.Consul.Address
		c.consulConfig.TLSConfig.CAFile = c.Config.Consul.CAFile
		c.consulConfig.TLSConfig.CertFile = c.Config.Consul.CertFile
		c.consulConfig.TLSConfig.KeyFile = c.Config.Consul.KeyFile
		c.consulConfig.TLSConfig.InsecureSkipVerify = c.Config.Consul.InsecureSkipVerify
	default:
		c.consulConfig.Address = c.Config.Consul.Address
	}

	httpClient, err := capi.NewHttpClient(c.consulConfig.Transport, c.consulConfig.TLSConfig)
	if err != nil {
		return nil, err
	}
	c.consulConfig.HttpClient = httpClient
	c.consulConfig.HttpClient.Timeout = c.Config.Consul.Timeout

	// Add Token
	if c.Config.Consul.Token != "" {
		c.consulConfig.Token = c.Config.Consul.Token
	}

	cc, err := capi.NewClient(
		c.consulConfig,
	)
	if err != nil {
		return nil, err
	}
	c.client = cc

	c.Logger.Debug("new Consul client has been initialized")
	return c, nil
}

func (c *Client) IsServiceExists(ctx *fiber.Ctx) (bool, error) {
	namespace := ctx.Params("namespace")
	service := ctx.Params("service")
	requestID := ctx.GetRespHeader("X-Request-Id")
	queryOptions := &capi.QueryOptions{}

	if namespace != "" {
		queryOptions.Namespace = namespace
	}

	s, _, err := c.client.Catalog().Service(service, ctx.Query("tag"), queryOptions)
	if err != nil {
		return false, err
	}

	if len(s) == 0 {
		c.Logger.Debug("Consul service doesn't exist",
			zap.String("request_id", requestID),
			zap.String("namespace", namespace),
			zap.String("service", service),
		)
		return false, nil
	}

	return true, nil
}

func (c *Client) IsServiceHealthy(ctx *fiber.Ctx) (bool, error) {
	namespace := ctx.Params("namespace")
	service := ctx.Params("service")
	queryOptions := &capi.QueryOptions{}

	if namespace != "" {
		queryOptions.Namespace = namespace
	}

	serviceEntry, _, err := c.client.Health().Service(service, ctx.Query("tag"), true, queryOptions)
	if err != nil {
		return false, err
	}

	if len(serviceEntry) == 0 {
		return false, nil
	}

	return true, nil
}

func (c *Client) Close() {
	c.consulConfig.HttpClient.CloseIdleConnections()
}

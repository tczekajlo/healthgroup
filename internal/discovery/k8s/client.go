package k8s

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/tczekajlo/healthgroup/internal/config"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Logger *zap.Logger
	Config *config.Config

	clientset  *kubernetes.Clientset
	httpClient *http.Client
}

func New(c *Client) (*Client, error) {
	if !c.Config.Kubernetes.Enabled {
		return nil, xerrors.New("Kubernetes client is disabled. You can enabled it in the configuration file")
	}

	var (
		config *rest.Config
		err    error
	)

	flags := c.Config.Flags()

	if flags.InCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", flags.Kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	// creates the clientset
	c.httpClient, err = rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfigAndClient(config, c.httpClient)
	if err != nil {
		return nil, err
	}
	c.clientset = clientset

	c.Logger.Debug("new Kubernetes client has been initialized")
	return c, nil
}

func (c *Client) IsServiceExists(ctx *fiber.Ctx) (bool, error) {
	namespace := ctx.Params("namespace")
	service := ctx.Params("service")
	requestID := ctx.GetRespHeader("X-Request-Id")

	_, err := c.clientset.CoreV1().Services(namespace).Get(context.TODO(), service, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		c.Logger.Debug("Kubernetes service doesn't exist",
			zap.String("request_id", requestID),
			zap.String("namespace", namespace),
			zap.String("service", service),
		)
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) GetEndpoint(namespace, service string) (*v1.Endpoints, error) {
	return c.clientset.CoreV1().Endpoints(namespace).Get(context.TODO(), service, metav1.GetOptions{})
}

func (c *Client) IsServiceHealthy(ctx *fiber.Ctx) (bool, error) {
	namespace := ctx.Params("namespace")
	service := ctx.Params("service")
	requestID := ctx.GetRespHeader("X-Request-Id")

	svcExists, err := c.IsServiceExists(ctx)
	if err != nil {
		return false, err
	}

	if svcExists {
		endpoint, err := c.GetEndpoint(namespace, service)
		if err != nil {
			return false, err
		}

		if len(endpoint.Subsets) > 0 {
			c.Logger.Debug("Kubernetes service is healthy",
				zap.String("request_id", requestID),
				zap.String("namespace", namespace),
				zap.String("service", service),
				zap.Any("endpoint_subnets", endpoint.Subsets),
			)
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}

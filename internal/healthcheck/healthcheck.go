package healthcheck

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tczekajlo/healthgroup/internal/config"
	"github.com/tczekajlo/healthgroup/internal/version"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

const (
	HTTP2 string = "http2"
	HTTP  string = "http"
)

type HealthCheck struct {
	Logger    *zap.Logger
	Config    *config.Config
	Discovery string
}

func (h *HealthCheck) Run(c *fiber.Ctx) error {
	g := new(errgroup.Group)

	g.Go(func() error {
		return h.runHTTPHealthCheck(c)
	})

	// Wait for all HTTP fetches to complete.
	return g.Wait()
}

func (h *HealthCheck) runHTTPHealthCheck(c *fiber.Ctx) error {
	g := new(errgroup.Group)
	g.SetLimit(h.Config.Concurrency)

	for _, check := range h.Config.HTTPHealthCheck {
		// Skip a given health check if service or namespace doesn't match.
		if h.shouldSkip(c, check) {
			continue
		}

		healthCheck := check // https://golang.org/doc/faq#closures_and_goroutines

		g.Go(func() error {
			return h.execHTTPHealthCheck(c, healthCheck)
		})
	}
	return g.Wait()
}

func (h *HealthCheck) shouldSkip(c *fiber.Ctx, check interface{}) bool {
	var checkNS, checkSVC, checkDiscovery string
	namespace := c.Params("namespace")
	service := c.Params("service")
	requestID := c.GetRespHeader("X-Request-Id")

	switch v := check.(type) { //nolint:gocritic
	case config.HTTPHealthCheck:
		checkNS = v.Namespace
		checkSVC = v.Service
		checkDiscovery = strings.ToLower(v.Discovery)
	}

	if h.Discovery != checkDiscovery && checkDiscovery != "" {
		h.Logger.Debug("skip health check",
			zap.String("request_id", requestID),
			zap.Any("health_check", check),
		)
		return true
	}

	if checkNS != namespace && checkNS != "" {
		h.Logger.Debug("skip health check",
			zap.String("request_id", requestID),
			zap.Any("health_check", check),
		)
		return true
	}

	if checkSVC != service && checkSVC != "" {
		h.Logger.Debug("skip health check",
			zap.String("request_id", requestID),
			zap.Any("health_check", check),
		)
		return true
	}

	return false
}

func (h *HealthCheck) execHTTPHealthCheck(c *fiber.Ctx, check config.HTTPHealthCheck) error {
	requestID := c.GetRespHeader("X-Request-Id")
	url, err := buildURL(check)
	if err != nil {
		h.Logger.Error("external health check",
			zap.String("request_id", requestID),
			zap.Any("health_check", check),
			zap.Error(err),
		)
		return err
	}

	client := http.Client{
		Timeout: check.Timeout * time.Second,
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: check.InsecureSkipVerify, //nolint:gosec
	}

	switch check.Type {
	case HTTP2:
		client.Transport = &http2.Transport{
			TLSClientConfig: tlsConfig,
		}
	default:
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("healthgroup/%s", version.Version))

	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
		client.CloseIdleConnections()

		h.Logger.Info("external health check",
			zap.String("request_id", requestID),
			zap.String("url", url),
			zap.Int("status", resp.StatusCode),
		)

		if resp.StatusCode != http.StatusOK {
			return xerrors.Errorf("health check failed, status code: %d, url: %s", resp.StatusCode, url)
		}
	} else {
		resp.Body.Close()
		client.CloseIdleConnections()
		h.Logger.Error("external health check",
			zap.String("request_id", requestID),
			zap.String("url", url),
			zap.Error(err),
		)
	}

	return err
}

func buildURL(healthCheck config.HTTPHealthCheck) (string, error) {
	var url string
	hType := strings.ToLower(healthCheck.Type)

	if hType != HTTP && hType != "https" && hType != HTTP2 {
		return "", xerrors.Errorf("health check type is not supported, type: %s", hType)
	}

	switch hType {
	case HTTP2:
		url = fmt.Sprintf("https://%s", healthCheck.Host)
	default:
		url = fmt.Sprintf("%s://%s", hType, healthCheck.Host)
	}

	if healthCheck.Port != 0 {
		url = fmt.Sprintf("%s:%d", url, healthCheck.Port)
	}

	url = fmt.Sprintf("%s%s", url, healthCheck.RequestPath)

	return url, nil
}

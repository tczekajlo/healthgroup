package discovery

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tczekajlo/healthgroup/internal/config"
	"go.uber.org/zap"
)

type Discovery struct {
	Logger *zap.Logger
	Config *config.Config
	Source string
}

type Adapter interface {
	IsServiceExists(ctx *fiber.Ctx) (bool, error)
	IsServiceHealthy(ctx *fiber.Ctx) (bool, error)
	Close()
}

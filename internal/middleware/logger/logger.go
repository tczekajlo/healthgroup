package logger

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Config defines the config for middleware
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool

	// Logger defines zap logger instance
	Logger *zap.Logger
}

// New creates a new middleware handler
func New(config Config) fiber.Handler {
	var (
		errPadding  = 15
		start, stop time.Time
		once        sync.Once
		errHandler  fiber.ErrorHandler
	)

	return func(c *fiber.Ctx) error {
		if config.Next != nil && config.Next(c) {
			return c.Next()
		}

		once.Do(func() {
			errHandler = c.App().Config().ErrorHandler
			stack := c.App().Stack()
			for m := range stack {
				for r := range stack[m] {
					if len(stack[m][r].Path) > errPadding {
						errPadding = len(stack[m][r].Path)
					}
				}
			}
		})

		start = time.Now()

		chainErr := c.Next()

		if chainErr != nil {
			if err := errHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		stop = time.Now()

		fields := []zap.Field{
			zap.String("request_id", c.GetRespHeader("X-Request-Id")),
			zap.Int64("duration_ms", stop.Sub(start).Milliseconds()),
			zap.Int("status", c.Response().StatusCode()),
			zap.String("url", c.OriginalURL()),
			zap.String("agent", string(c.Context().UserAgent())),
			zap.String("ip", c.IP()),
			zap.String("method", c.Method()),
		}

		config.Logger.With(fields...).Info("request")

		return nil
	}
}

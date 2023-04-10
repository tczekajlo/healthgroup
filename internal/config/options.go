package config

import "go.uber.org/zap"

func WithLogger(logger *zap.Logger) Option {
	return func(c *Config) {
		c.logger = logger
	}
}

func WithFile(file string) Option {
	return func(c *Config) {
		c.file = file
	}
}

func WithFlags(flags *Flags) Option {
	return func(c *Config) {
		c.flags = flags
	}
}

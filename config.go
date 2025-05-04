package pal

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	configValidator = validator.New()
)

// Config is the configuration for pal.
type Config struct {
	InitTimeout        time.Duration `Validate:"gt=0"`
	HealthCheckTimeout time.Duration `Validate:"gt=0"`
	ShutdownTimeout    time.Duration `Validate:"gt=0"`
}

func (c *Config) Validate(_ context.Context) error {
	return configValidator.Struct(c)
}

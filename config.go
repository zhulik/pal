package pal

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

// Config is the configuration for pal.
type Config struct {
	InitTimeout        time.Duration `validate:"gt=0"`
	HealthCheckTimeout time.Duration `validate:"gt=0"`
	ShutdownTimeout    time.Duration `validate:"gt=0"`
}

func (c *Config) validate(_ context.Context) error {
	validate := validator.New()
	return validate.Struct(c)
}

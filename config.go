package pal

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	configValidator = validator.New()
)

// SlogAttributeSetter is a function that returns the name and value for the attribute to be added to the slog.Logger.
// It receives the target struct and returns the name and value for the attribute. Called when logger is being injected.
type SlogAttributeSetter func(target any) (string, string)

// Config is the configuration for pal.
type Config struct {
	InitTimeout        time.Duration `validate:"gt=0"`
	HealthCheckTimeout time.Duration `validate:"gt=0"`
	ShutdownTimeout    time.Duration `validate:"gt=0"`
	InjectSlog         bool
	AttrSetters        []SlogAttributeSetter
}

func (c *Config) Validate(_ context.Context) error {
	return configValidator.Struct(c)
}

package pal

import (
	"context"
	"time"
)

// Config is the configuration for pal.
type Config struct {
	InitTimeout        time.Duration // required
	HealthCheckTimeout time.Duration // required
	ShutdownTimeout    time.Duration // required
}

func (c *Config) validate(_ context.Context) error {
	// TODO: write me
	return nil
}

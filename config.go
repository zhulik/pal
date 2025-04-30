package pal

import "time"

// Config is the configuration for pal.
// TODO: validate config with validator?
type Config struct {
	InitTimeout        time.Duration // required
	HealthCheckTimeout time.Duration // required
	ShutdownTimeout    time.Duration // required
}

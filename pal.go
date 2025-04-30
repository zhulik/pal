package pal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ContextKey int

const (
	CtxValue ContextKey = iota
)

type Pal struct {
	config *Config

	factories []DependencyFactory
}

func New(factories ...DependencyFactory) *Pal {
	return &Pal{
		config:    &Config{},
		factories: factories,
	}
}

// Provide registers a singleton service with pal. *I* must be an interface and *S* must be a struct that implements I.
// Only one instance of the service will be created and reused.
// TODO: any ways to enforce this with types?
func Provide[I any, S Service]() *Dependency[I, S] {
	return &Dependency[I, S]{
		singleton: true,
	}
}

// ProvideFactory registers a factory service with pal. *I* must be an interface, and *S* must be a struct that implements I.
// A new factory service instances are created every time the service is invoked.
// it's the caller's responsibility to shut down the service, pal will also not healthcheck it.
func ProvideFactory[I any, S Service]() *Dependency[I, S] {
	return &Dependency[I, S]{
		singleton: false,
	}
}

// InitTimeout sets the timeout for the initialization of the services.
func (p *Pal) InitTimeout(t time.Duration) *Pal {
	p.config.InitTimeout = t
	return p
}

// HealthCheckTimeout sets the timeout for the healthcheck of the services.
func (p *Pal) HealthCheckTimeout(t time.Duration) *Pal {
	p.config.HealthCheckTimeout = t
	return p
}

// ShutdownTimeout sets the timeout for the shutdown of the services.
func (p *Pal) ShutdownTimeout(t time.Duration) *Pal {
	p.config.ShutdownTimeout = t
	return p
}

// Error triggers graceful shutdown of the app, the error will be printer out, Pal.Run() will return an error.
func (p *Pal) Error(_ error) {
}

func (p *Pal) Shutdown(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
	defer cancel()
	return nil
}

func (p *Pal) validate(_ context.Context) error {
	// TODO: validate config here
	return nil
}

func (p *Pal) init(_ context.Context) error {
	// TODO: go through all the factories and create runners
	return nil
}

// Run eagerly initializes and starts Runners, then blocks until one of the given signals is received.
// When it's received, pal will gracefully shut down the app.
func (p *Pal) Run(ctx context.Context, _ ...syscall.Signal) error {
	err := p.validate(ctx)
	if err != nil {
		return err
	}

	err = p.init(ctx)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		return &RunError{ctx.Err()}
	case <-sigChan:
		return &RunError{p.Shutdown(ctx)}
	}
}

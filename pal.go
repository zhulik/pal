package pal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

type ContextKey int

const (
	CtxValue ContextKey = iota
)

type Pal struct {
	config *Config
	store  *store

	log loggerFn
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

// SetLogger sets the logger instance to be used by Pal
func (p *Pal) SetLogger(log loggerFn) *Pal {
	p.log = log
	return p
}

// Error triggers graceful shutdown of the app, the error will be printer out, Pal.Run() will return an error.
func (p *Pal) Error(_ error) {
	// TODO: write me
}

func (p *Pal) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.config.HealthCheckTimeout)
	defer cancel()

	err := p.store.healthCheck(ctx)
	if err != nil {
		p.Error(err)
	}
	return nil
}

// Shutdown gracefully stops all services within the configured timeout duration. Returns an error if shutdown fails.
func (p *Pal) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
	defer cancel()

	return p.store.shutdown(ctx)
}

// Run eagerly initializes and starts Runners, then blocks until one of the given signals is received.
// When it's received, pal will gracefully shut down the app.
func (p *Pal) Run(ctx context.Context, signals ...os.Signal) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	if err := p.validate(ctx); err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	err := p.store.init(initCtx, p)
	cancel()
	if err != nil {
		p.Error(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	p.log("running until one of %+v is received", signals)

	select {
	case <-ctx.Done():
		return &RunError{ctx.Err()}
	case <-sigChan:
		if err := p.Shutdown(ctx); err != nil {
			return &RunError{p.Shutdown(ctx)}
		}
		return nil
	}
}

func (p *Pal) Services() []string {
	return p.store.services()
}

func (p *Pal) Runners() []string {
	return p.store.runners()
}

func (p *Pal) Invoke(ctx context.Context, name string) (any, error) {
	ctx = context.WithValue(ctx, CtxValue, p)
	p.log("invoking %s", name)

	factory, ok := p.store.factories[name]
	if !ok {
		return nil, fmt.Errorf("%w: '%s', known services: %s", ErrServiceNotFound, name, p.Services())
	}

	var instance any

	if factory.IsSingleton() {
		instance, ok = p.store.instances[name]
		if !ok {
			return nil, fmt.Errorf("%w: '%s'", ErrServiceNotInit, name)
		}
	} else {
		var err error
		instance, err = factory.Initialize(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: '%s'", ErrServiceInitFailed, name)
		}
	}

	return instance, nil
}

func (p *Pal) validate(_ context.Context) error {
	// TODO: write me
	// TODO: validate config here
	return nil
}

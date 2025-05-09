package pal

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"
)

type ContextKey int

const (
	CtxValue ContextKey = iota
)

type Pal struct {
	config    *Config
	container *Container

	// stopChan is used to initiate the shutdown of the app.
	stopChan chan error

	// shutdownChan is used to wait for the graceful shutdown of the app.
	shutdownChan chan error

	initialized bool

	logger Logger
}

// New creates and returns a new instance of Pal with the provided Service's
func New(services ...ServiceImpl) *Pal {
	return &Pal{
		config:       &Config{},
		container:    NewContainer(services...),
		stopChan:     make(chan error, 1),
		shutdownChan: make(chan error, 1),
		logger:       emptyLogger,
	}
}

// FromContext retrieves a *Pal from the provided context, expecting it to be stored under the CtxValue key.
// Panics if ctx misses the value.
func FromContext(ctx context.Context) Context {
	return ctx.Value(CtxValue).(Context)
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

// ShutdownTimeout sets the timeout for the Shutdown of the services.
func (p *Pal) ShutdownTimeout(t time.Duration) *Pal {
	p.config.ShutdownTimeout = t
	return p
}

// SetLogger sets the logger instance to be used by Pal
func (p *Pal) SetLogger(logger Logger) *Pal {
	p.logger = logger
	p.container.SetLogger(logger)
	return p
}

// HealthCheck verifies the health of the service Container within a configurable timeout.
func (p *Pal) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.config.HealthCheckTimeout)
	defer cancel()

	return p.container.HealthCheck(ctx)
}

// Shutdown schedules graceful Shutdown of the app. If any errs given - Run() will return them. Only the first call is effective.
// The later calls are ignored.
func (p *Pal) Shutdown(errs ...error) {
	err := errors.Join(errs...)

	select {
	case p.stopChan <- err:
	default:
		if err != nil {
			p.logger.Warn("Shutdown already scheduled", "error", err)
		}
	}
}

// Run eagerly starts runners, then blocks until one of the given signals is received or all runners
// finish their work. If any error occurs during initialization, runner operation or Shutdown - Run() will return it.
func (p *Pal) Run(ctx context.Context, signals ...os.Signal) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	if err := p.Init(ctx); err != nil {
		return err
	}

	p.logger.Info("Pal initialized", "services", p.Services())

	go p.listenToStopSignals(ctx, signals)
	go func() {
		p.Shutdown(p.container.StartRunners(ctx))
	}()

	p.logger.Info("Running until signal is received or until job is done", "signals", signals)

	return <-p.shutdownChan
}

// Init initializes Pal. Validates config, creates and initializes all singleton services.
func (p *Pal) Init(ctx context.Context) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	if p.initialized {
		return nil
	}

	if err := p.validate(ctx); err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	defer cancel()

	if err := p.container.Init(initCtx); err != nil {
		p.logger.Warn("Init failed", "error", err)

		p.Shutdown(err)
		return err
	}

	p.initialized = true

	go func() {
		err := <-p.stopChan

		p.logger.Warn("Shutdown requested", "error", err)

		go func() {
			<-time.After(p.config.ShutdownTimeout)

			panic("shutdown timed out")
		}()

		shutCt, cancel := context.WithTimeout(ctx, time.Duration(float64(p.config.ShutdownTimeout)*0.9))
		defer cancel()

		p.shutdownChan <- errors.Join(err, p.container.Shutdown(shutCt))
		p.logger.Info("Shutdown completed.")
	}()

	return nil
}

func (p *Pal) Services() []ServiceImpl {
	return p.container.Services()
}

func (p *Pal) Invoke(ctx context.Context, name string) (any, error) {
	ctx = context.WithValue(ctx, CtxValue, p)
	p.logger.Debug("Invoking", "service", name)

	return p.container.Invoke(ctx, name)
}

func (p *Pal) validate(ctx context.Context) error {
	return errors.Join(p.config.Validate(ctx), p.container.Validate(ctx))
}

func (p *Pal) listenToStopSignals(ctx context.Context, signals []os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	select {
	case <-ctx.Done():
		p.Shutdown(ctx.Err())
	case sig := <-sigChan:
		p.logger.Warn("Received signal", "signal", sig)

		p.Shutdown()
	}
}

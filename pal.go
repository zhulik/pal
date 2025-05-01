package pal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

type ContextKey int

const (
	CtxValue ContextKey = iota
)

type Pal struct {
	config   *Config
	store    *store
	stopChan chan error

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

func (p *Pal) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.config.HealthCheckTimeout)
	defer cancel()

	err := p.store.healthCheck(ctx)
	if err != nil {
		p.Shutdown(err)
	}
	return err
}

// Shutdown schedules graceful shutdown of the app. if any errs given - Run() will return them. Only the first call is effective.
// The later calls are queued but ignored.
func (p *Pal) Shutdown(errs ...error) {
	// In theory this causes a goroutine leak, but it's not a big deal as we are shutting down anyway.
	go func() {
		p.stopChan <- errors.Join(errs...)
	}()
}

// Run eagerly initializes and starts Runners, then blocks until one of the given signals is received or all runners
// finish their work. If any error occurs during initialization, runner operation or shutdown - Run() will return it.
func (p *Pal) Run(ctx context.Context, signals ...os.Signal) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	if err := p.validate(ctx); err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	defer cancel()

	if err := p.store.init(initCtx); err != nil {
		shutCtx, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
		defer cancel()
		return errors.Join(err, p.store.shutdown(shutCtx))
	}

	p.startRunners(ctx)

	go forwardSignals(signals, p.stopChan)

	go func() {
		<-ctx.Done()
		p.stopChan <- ctx.Err()
	}()

	p.log("running until one of %+v is received or until job is done", signals)

	err := <-p.stopChan

	shutCt, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
	defer cancel()
	return errors.Join(err, p.store.shutdown(shutCt))
}

func forwardSignals(signals []os.Signal, ch chan error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	<-sigChan
	ch <- nil
}

func (p *Pal) startRunners(ctx context.Context) {
	g := &errgroup.Group{}

	for _, name := range p.Runners() {
		g.Go(func() error {
			p.log("running %s", name)
			err := p.store.instances[name].(Runner).Run(ctx)
			defer p.log("%s exited with error='%+v'", name, err)

			return err
		})
	}

	go func() {
		p.Shutdown(g.Wait())
	}()
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
		p.log("initializing %s", name)
		instance, err = factory.Initialize(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: '%s'", ErrServiceInitFailed, name)
		}
	}

	return instance, nil
}

func (p *Pal) validate(ctx context.Context) error {
	errs := []error{p.config.validate(ctx)}

	for _, factory := range p.store.factories {
		if err := factory.Validate(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

package pal

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/zhulik/pal/internal/container"
	"github.com/zhulik/pal/pkg/core"

	"golang.org/x/sync/errgroup"
)

type ContextKey int

const (
	CtxValue ContextKey = iota
)

type Pal struct {
	config   *core.Config
	store    *container.Container
	stopChan chan error

	log core.LoggerFn
}

// New creates and returns a new instance of Pal with the provided Service's
func New(services ...core.Service) *Pal {
	index := make(map[string]core.Service)

	for _, service := range services {
		index[service.Name()] = service
	}

	logger := func(string, ...any) {}

	return &Pal{
		config:   &core.Config{},
		store:    container.New(index),
		stopChan: make(chan error),
		log:      logger,
	}
}

// FromContext retrieves a *Pal from the provided context, expecting it to be stored under the CtxValue key.
// Panics if ctx misses the value.
func FromContext(ctx context.Context) *Pal {
	return ctx.Value(CtxValue).(*Pal)
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
func (p *Pal) SetLogger(log core.LoggerFn) *Pal {
	p.log = log
	p.store.SetLogger(log)
	return p
}

// HealthCheck verifies the health of the service Container within a configurable timeout.
func (p *Pal) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.config.HealthCheckTimeout)
	defer cancel()

	return p.store.HealthCheck(ctx)
}

// Shutdown schedules graceful Shutdown of the app. If any errs given - Run() will return them. Only the first call is effective.
// The later calls are ignored.
func (p *Pal) Shutdown(errs ...error) {
	// In theory this causes a goroutine leak, but it's not a big deal as we are shutting down anyway.
	// TODO: figure out how to handle multiple calls to Shutdown.
	go func() {
		p.stopChan <- errors.Join(errs...)
	}()
}

// Run eagerly initializes and starts Runners, then blocks until one of the given signals is received or all Runners
// finish their work. If any error occurs during initialization, runner operation or Shutdown - Run() will return it.
func (p *Pal) Run(ctx context.Context, signals ...os.Signal) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	// TODO: skip if already
	if err := p.Init(ctx); err != nil {
		return err
	}

	p.log("Pal initialized. Services: %s", p.Services())

	p.startRunners(ctx)

	go p.forwardSignals(signals)

	go func() {
		<-ctx.Done()
		p.stopChan <- ctx.Err()
	}()

	p.log("running until one of %+v is received or until job is done", signals)

	err := <-p.stopChan

	shutCt, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
	defer cancel()
	return errors.Join(err, p.store.Shutdown(shutCt))
}

func (p *Pal) Init(ctx context.Context) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	if err := p.validate(ctx); err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	defer cancel()

	if err := p.store.Init(initCtx); err != nil {
		p.log("Init failed with %+v", err)

		shutCtx, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
		defer cancel()

		return errors.Join(err, p.store.Shutdown(shutCtx))
	}
	return nil
}

func (p *Pal) Services() []core.Service {
	return p.store.Services()
}

func (p *Pal) Invoke(ctx context.Context, name string) (any, error) {
	ctx = context.WithValue(ctx, CtxValue, p)
	p.log("invoking %s", name)

	return p.store.Invoke(ctx, name)
}

func (p *Pal) validate(ctx context.Context) error {
	return errors.Join(p.config.Validate(ctx), p.store.Validate(ctx))
}

func (p *Pal) forwardSignals(signals []os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	sig := <-sigChan

	p.log("signal received: %+v", sig)

	p.stopChan <- nil
}

func (p *Pal) startRunners(ctx context.Context) {
	g := &errgroup.Group{}

	for name, runner := range p.store.Runners(ctx) {
		g.Go(func() error {
			p.log("running %s", name)
			err := runner.Run(ctx)
			if err != nil {
				p.log("runner %s exited with error='%+v'", name, err)
				return err
			}

			p.log("runner %s finished successfully", name)
			return nil
		})
	}

	go p.Shutdown(g.Wait())
}

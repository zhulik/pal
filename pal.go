package pal

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

// ContextKey is a type used for context value keys to avoid collisions.
type ContextKey int

const (
	// CtxValue is the key used to store and retrieve the Pal instance from a context.
	// This allows services to access the Pal instance from a context passed to them.
	CtxValue ContextKey = iota
)

// Pal is the main struct that manages the lifecycle of services in the application.
// It handles service initialization, dependency injection, health checking, and graceful shutdown.
// Pal implements the Invoker interface, allowing services to be retrieved from it.
type Pal struct {
	config    *Config
	container *Container

	// stopChan is used to initiate the shutdown of the app.
	stopChan chan error

	// shutdownChan is used to wait for the graceful shutdown of the app.
	shutdownChan chan error

	initialized bool

	logger *slog.Logger
}

// New creates and returns a new instance of Pal with the provided Services
func New(services ...ServiceDef) *Pal {
	pal := &Pal{
		config:       &Config{},
		stopChan:     make(chan error, 1),
		shutdownChan: make(chan error, 1),
		logger:       slog.With("palComponent", "Pal"),
	}

	services = append(services, ProvideConst(pal))

	for _, s := range services {
		setPalField(reflect.ValueOf(s), pal)
	}

	pal.container = NewContainer(services...)

	return pal
}

func setPalField(v reflect.Value, pal *Pal) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.CanSet() && field.Type() == reflect.TypeOf(pal) {
			field.Set(reflect.ValueOf(pal))
		}

		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Pointer && !field.IsNil()) {
			setPalField(field, pal)
		}

		if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				item := field.Index(i)
				setPalField(item, pal)
			}
		}
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
			p.logger.Error("Shutdown already scheduled", "error", err)
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

	go p.listenToStopSignals(ctx, signals)
	go p.Shutdown(p.container.StartRunners(ctx))

	p.logger.Info("Running until signal is received or until job is done", "signals", signals)

	return <-p.shutdownChan
}

// Init initializes Pal. Validates config, creates and initializes all singleton services.
func (p *Pal) Init(ctx context.Context) error {
	if p.initialized {
		return nil
	}

	ctx = context.WithValue(ctx, CtxValue, p)

	if err := p.validate(ctx); err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	defer cancel()

	if err := p.container.Init(initCtx); err != nil {
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
	}()

	return nil
}

// Services returns a map of all registered services in the container, keyed by their names.
// This can be useful for debugging or introspection purposes.
func (p *Pal) Services() map[string]ServiceDef {
	return p.container.Services()
}

// Invoke retrieves a service by name from the container.
// It implements the Invoker interface.
// The context is enriched with the Pal instance before being passed to the container.
func (p *Pal) Invoke(ctx context.Context, name string) (any, error) {
	ctx = context.WithValue(ctx, CtxValue, p)

	return p.container.Invoke(ctx, name)
}

// InjectInto injects services into the fields of the target struct.
// It implements the Invoker interface.
// The context is enriched with the Pal instance before being passed to the container.
func (p *Pal) InjectInto(ctx context.Context, target any) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	return p.container.InjectInto(ctx, target)
}

// Container returns the underlying Container instance.
// This can be useful for advanced use cases where direct access to the container is needed.
func (p *Pal) Container() *Container {
	return p.container
}

// Logger returns the logger instance used by Pal.
// This can be useful for advanced use cases where direct access to the logger is needed.
func (p *Pal) Logger() *slog.Logger {
	return p.logger
}

func (p *Pal) validate(ctx context.Context) error {
	return errors.Join(p.config.Validate(ctx), p.container.Validate(ctx))
}

func (p *Pal) listenToStopSignals(ctx context.Context, signals []os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	shuttingDown := false

	for {
		select {
		case <-ctx.Done():
			p.Shutdown()
		case sig := <-sigChan:
			if shuttingDown {
				if sig == syscall.SIGINT {
					p.logger.Error("Signal received again, exiting immediately", "signal", sig)
					os.Exit(1)
				}
				p.logger.Warn("Received signal, but the app is already shutting down, ignoring.", "signal", sig)
				continue
			}

			p.logger.Warn("Received signal, shutting down.", "signal", sig)
			if sig == syscall.SIGINT {
				p.logger.Warn("Received signal, shutting down. Press ^C again to exit immediately", "signal", sig)
			} else {
				p.logger.Warn("Received signal, shutting down.", "signal", sig)
			}
			shuttingDown = true

			p.Shutdown()
		}
	}
}

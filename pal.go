package pal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
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

// DefaultShutdownSignals is the default signals that will be used to shutdown the app.
var DefaultShutdownSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

var defaultAttrSetters = []SlogAttributeSetter{
	func(target any) (string, string) {
		return "component", fmt.Sprintf("%T", target)
	},
}

// Pal is the main struct that manages the lifecycle of services in the application.
// It handles service initialization, dependency injection, health checking, and graceful shutdown.
// Pal implements the Invoker interface, allowing services to be retrieved from it.
type Pal struct {
	config    *Config
	container *Container

	initialized *atomic.Bool
	shutDown    *atomic.Bool

	logger *slog.Logger
}

// New creates and returns a new instance of Pal with the provided Services
func New(services ...ServiceDef) *Pal {
	pal := &Pal{
		config:      &Config{},
		initialized: &atomic.Bool{},
		shutDown:    &atomic.Bool{},
		logger:      slog.With("palComponent", "Pal"),
	}

	services = append(services, Provide(pal))

	for _, s := range services {
		setPalField(reflect.ValueOf(s), pal)
	}

	pal.container = NewContainer(pal.config, services...)

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

// InjectSlog enables automatic slog injection into the services.
func (p *Pal) InjectSlog(configs ...SlogAttributeSetter) *Pal {
	if len(configs) == 0 {
		configs = defaultAttrSetters
	}

	p.config.AttrSetters = configs
	return p
}

// RunHealthCheckServer enables the default health check server.
func (p *Pal) RunHealthCheckServer(addr, path string) *Pal {
	p.config.HealthCheckAddr = addr
	p.config.HealthCheckPath = path

	server := Provide[palHealthCheckServer](&healthCheckServer{})
	setPalField(reflect.ValueOf(server), p)
	p.container.services[server.Name()] = server

	return p
}

// HealthCheck verifies the health of the service Container within a configurable timeout.
func (p *Pal) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.config.HealthCheckTimeout)
	defer cancel()

	return p.container.HealthCheck(ctx)
}

// Init initializes Pal. Validates config, creates and initializes all singleton services.
// If any error occurs during initialization, it will return it and
// will not try to gracefully shutdown already initialized services.
// Only first call is effective.
func (p *Pal) Init(ctx context.Context) error {
	if !p.initialized.CompareAndSwap(false, true) {
		return nil
	}

	ctx = context.WithValue(ctx, CtxValue, p)

	if err := p.config.Validate(ctx); err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	defer cancel()

	if err := p.container.Init(initCtx); err != nil {
		return err
	}

	p.logger.Debug("Pal initialized")

	return nil
}

// shutdown gracefully shuts down the app. If any errs given or an error occurs during services shutdown, Run()
// will return them. If not initialized, the call is ignored.
func (p *Pal) shutdown() error {
	go func() {
		// a watchdog to make sure to forcefully exit if shutdown times out
		<-time.After(p.config.ShutdownTimeout)

		panic("shutdown timed out")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(float64(p.config.ShutdownTimeout)*0.9))
	defer cancel()

	return p.container.Shutdown(ctx)
}

// Run eagerly starts runners, then blocks until:
// - context is canceled
// - one of the runners fails
// - one of the given signals is received, if a signal is received again, the app will exit immediately without graceful shutdown
// - all runners finish their work
// Not goroutine safe, must only be called once.
// After one of the events above occurs, the app will be gracefully shot down.
// Errors returned from runners and during shutdown are collected and returned from Run().
func (p *Pal) Run(ctx context.Context, signals ...os.Signal) error {
	if len(signals) == 0 {
		signals = DefaultShutdownSignals
	}

	ctx = context.WithValue(ctx, CtxValue, p)

	ctx, stop := signal.NotifyContext(ctx, signals...)
	defer stop()

	if err := p.Init(ctx); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		p.logger.Warn("Received signal, shutting down.")

		ctx, stop := signal.NotifyContext(context.Background(), signals...)
		defer stop()

		<-ctx.Done()
		p.logger.Error("Signal received again, exiting immediately")
		os.Exit(1)
	}()

	p.logger.Info("Running until signal is received or until job is done", "signals", signals)
	runErr := p.container.StartRunners(ctx)

	if errors.Is(runErr, context.Canceled) || errors.Is(runErr, errNoMainRunners) {
		runErr = nil
	}

	if runErr != nil {
		p.logger.Error("One or more runners failed, trying to shutdown gracefully", "error", runErr)
	}

	return errors.Join(runErr, p.shutdown())
}

// Services returns a map of all registered services in the container, keyed by their names.
// This can be useful for debugging or introspection purposes.
func (p *Pal) Services() map[string]ServiceDef {
	return p.container.Services()
}

// Invoke retrieves a service by name from the container.
// It implements the Invoker interface.
// The context is enriched with the Pal instance before being passed to the container.
func (p *Pal) Invoke(ctx context.Context, name string, args ...any) (any, error) {
	ctx = context.WithValue(ctx, CtxValue, p)

	return p.container.Invoke(ctx, name, args...)
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

// Config returns a copy of pal's config.
func (p *Pal) Config() Config {
	return *p.config
}

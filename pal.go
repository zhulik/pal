package pal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"slices"
	"syscall"
	"time"

	"github.com/dominikbraun/graph"
)

type ContextKey int

const (
	CtxValue ContextKey = iota
)

type Pal struct {
	config *Config

	factories map[string]ServiceFactory
	graph     graph.Graph[string, string]
	instances map[string]any
}

func New(factories ...ServiceFactory) *Pal {
	index := make(map[string]ServiceFactory)

	for _, factory := range factories {
		index[factory.Name()] = factory
	}

	return &Pal{
		config:    &Config{},
		factories: index,
		instances: map[string]any{},
		graph:     graph.New(graph.StringHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}
}

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

// ShutdownTimeout sets the timeout for the shutdown of the services.
func (p *Pal) ShutdownTimeout(t time.Duration) *Pal {
	p.config.ShutdownTimeout = t
	return p
}

// Error triggers graceful shutdown of the app, the error will be printer out, Pal.Run() will return an error.
func (p *Pal) Error(_ error) {
	// TODO: write me
}

func (p *Pal) Shutdown(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
	defer cancel()
	// TODO: write me

	return nil
}

// Run eagerly initializes and starts Runners, then blocks until one of the given signals is received.
// When it's received, pal will gracefully shut down the app.
func (p *Pal) Run(ctx context.Context, _ ...syscall.Signal) error {
	ctx = context.WithValue(ctx, CtxValue, p)

	if err := p.validate(ctx); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, CtxValue, p)
	ctx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)

	err := p.init(ctx)
	cancel()
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

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
	var names []string
	for name := range p.factories {
		names = append(names, name)
	}
	return names
}

func (p *Pal) Runners() []string {
	var runners []string
	for name, factory := range p.factories {
		if factory.IsRunner() {
			runners = append(runners, name)
		}
	}
	return runners
}

func (p *Pal) Invoke(ctx context.Context, name string) (any, error) {
	ctx = context.WithValue(ctx, CtxValue, p)

	factory, ok := p.factories[name]
	if !ok {
		return nil, fmt.Errorf("%w: '%s', known services: %s", ErrServiceNotFound, name, p.Services())
	}

	var instance any
	var err error

	if factory.IsSingleton() {
		instance, ok = p.instances[name]
		if !ok {
			return nil, fmt.Errorf("%w: '%s'", ErrServiceNotInit, name)
		}
	} else {
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

func (p *Pal) init(ctx context.Context) error {
	defer func() {
		// TODO: if init fails with an error - try to gracefully shutdown already initialized dependencies.
	}()

	// Initialize dependencies staring from the leaves. The more dependants an services, the earlier it will be initialized.

	err := p.buildDAG()
	if err != nil {
		return err
	}

	// file, _ := os.Initialize("./mygraph.gv")
	// _ = draw.DOT(p.graph, file)

	order, err := graph.TopologicalSort(p.graph)
	if err != nil {
		return err
	}
	slices.Reverse(order)

	for _, factoryName := range order {
		factory := p.factories[factoryName]
		if !factory.IsSingleton() {
			continue
		}

		instance, err := factory.Initialize(ctx)
		if err != nil {
			return err
		}

		p.instances[factoryName] = instance
	}

	for _, instance := range p.instances {
		if runner, ok := instance.(Runner); ok {
			go func() {
				// TODO: use a custom context struct?
				ctx := context.WithValue(context.Background(), CtxValue, p)
				err := runner.Run(ctx)

				if err != nil {
					p.Error(err)
				}
			}()
		}
	}
	return nil
}

func (p *Pal) buildDAG() error {
	runners := p.Runners()

	log.Printf("deps: %s", p.Services())
	log.Printf("runners: %s", runners)

	for _, runner := range runners {
		err := p.addDependencyVertex(runner, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pal) addDependencyVertex(name string, parent string) error {
	if err := p.graph.AddVertex(name); err != nil {
		if !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return err
		}
	}

	if parent != "" {
		if err := p.graph.AddEdge(parent, name); err != nil {
			return err
		}
	}

	factory := p.factories[name]

	instance := factory.Make()

	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Type.Kind() == reflect.Interface {
			dependencyName := field.Type.String()
			if _, ok := p.factories[dependencyName]; ok {
				if err := p.addDependencyVertex(dependencyName, name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

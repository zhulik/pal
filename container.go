package pal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/zhulik/pal/pkg/pid"

	"golang.org/x/sync/errgroup"

	"github.com/zhulik/pal/pkg/dag"
)

// Container is responsible for storing services, instances and the dependency graph
type Container struct {
	config *Config

	services map[string]ServiceDef
	graph    *dag.DAG[string, ServiceDef]
	logger   *slog.Logger

	runnerTasks *pid.RunGroup

	cancelMu sync.RWMutex
	cancel   context.CancelFunc
}

// NewContainer creates a new Container instance
func NewContainer(config *Config, services ...ServiceDef) *Container {
	services = flattenServices(services)
	index := make(map[string]ServiceDef)

	for _, service := range services {
		index[service.Name()] = service
	}

	return &Container{
		config:      config,
		services:    index,
		graph:       dag.New(serviceHash),
		logger:      slog.With("palComponent", "Container"),
		runnerTasks: pid.NewRunGroup(),
	}
}

func flattenServices(services []ServiceDef) []ServiceDef {
	seen := make(map[ServiceDef]bool)
	var result []ServiceDef

	var process func([]ServiceDef)
	process = func(svcs []ServiceDef) {
		for _, svc := range svcs {
			if _, ok := seen[svc]; !ok {
				seen[svc] = true

				if !strings.HasPrefix(svc.Name(), "$") {
					result = append(result, svc)
				}

				process(svc.Dependencies())
			}
		}
	}

	process(services)
	return result
}

func (c *Container) Init(ctx context.Context) error {
	c.logger.Debug("Building dependency tree...")

	for _, service := range c.services {
		if err := c.addDependencyVertex(service, nil); err != nil {
			return err
		}
	}

	order, err := c.graph.ReverseTopologicalOrder()
	if err != nil {
		return err
	}

	c.logger.Debug("Dependency tree is built", "order", order)

	err = c.graph.InReverseTopologicalOrder(func(service ServiceDef) error {
		if err := service.Init(ctx); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.logger.Error("Failed to initialize container", "error", err)
		return err
	}

	c.logger.Debug("Container initialized")
	return nil
}

func (c *Container) Invoke(ctx context.Context, name string, args ...any) (any, error) {
	service, ok := c.services[name]
	if !ok {
		return nil, fmt.Errorf("%w: '%s', known services: %s", ErrServiceNotFound, name, c.services)
	}

	if len(args) != service.Arguments() {
		return nil, fmt.Errorf("%w: '%s': %d arguments expected, got %d", ErrServiceInvalidArgumentsCount, name, service.Arguments(), len(args))
	}

	instance, err := service.Instance(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s': %w", ErrServiceInitFailed, name, err)
	}

	return instance, nil
}

func (c *Container) InjectInto(ctx context.Context, target any) error {
	v := reflect.ValueOf(target).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)

		if !field.CanSet() {
			continue
		}

		fieldType := t.Field(i).Type
		if fieldType == reflect.TypeOf((*slog.Logger)(nil)) && c.config.AttrSetters != nil {
			c.injectLoggerIntoField(field, target)
			continue
		}

		dependency, err := c.Invoke(ctx, fieldType.String())
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			if errors.Is(err, ErrServiceInvalidArgumentsCount) {
				return fmt.Errorf("%w: '%s': %w", ErrFactoryServiceDependency, fieldType.String(), err)
			}
			return err
		}

		field.Set(reflect.ValueOf(dependency))
	}
	return nil
}

func (c *Container) Shutdown(ctx context.Context) error {
	var errs []error

	c.logger.Debug("Shutting down runners")
	// Shutting down runners by cancelling their root context
	c.cancelMu.RLock()
	if c.cancel != nil {
		c.cancel()
	}
	c.cancelMu.RUnlock()

	errs = append(errs, c.runnerTasks.Wait())
	// Await for runners to exit and save possible error.
	err := errors.Join(errs...)

	if err != nil {
		c.logger.Error("Runners failed to shutdown", "error", err)
	}

	c.graph.InTopologicalOrder(func(service ServiceDef) error { // nolint:errcheck
		err := service.Shutdown(ctx)
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		return nil
	})

	err = errors.Join(errs...)
	if err != nil {
		c.logger.Error("Failed to shutdown container", "error", err)
		return err
	}

	c.logger.Debug("Container shut down successfully")
	return nil
}

func (c *Container) HealthCheck(ctx context.Context) error {
	var wg errgroup.Group

	c.logger.Debug("Healthchecking services")

	c.graph.ForEachVertex(func(service ServiceDef) error { // nolint:errcheck
		wg.Go(func() error {
			// Do not check pal again, this leads to recursion
			if service.Name() == "*pal.Pal" {
				return nil
			}

			err := service.HealthCheck(ctx)
			if err != nil {
				return err
			}

			return nil
		})

		return nil
	})

	err := wg.Wait()
	if err != nil {
		c.logger.Error("Healthcheck failed", "error", err)
		return err
	}

	c.logger.Debug("Healthcheck successful")

	return nil
}

// Services returns a map of all registered services in the container, keyed by their names.
// This can be useful for debugging or introspection purposes.
func (c *Container) Services() map[string]ServiceDef {
	return c.services
}

// StartRunners starts all services that implement the Runner interface in background goroutines.
// It creates a cancellable context that will be canceled during shutdown.
// Returns an error if any runner fails, though runners continue to execute independently.
func (c *Container) StartRunners(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c.cancelMu.Lock()
	c.cancel = cancel
	c.cancelMu.Unlock()

	for _, service := range c.services {
		runCfg := runConfigOrDefault(service.RunConfig())

		c.runnerTasks.Go(ctx, runCfg.Wait, func(ctx context.Context) error {
			return service.Run(ctx)
		})
	}

	err := c.runnerTasks.Wait()
	if err != nil {
		c.logger.Error("Runners finished with error", "error", err)
		return nil
	}
	c.logger.Debug("All runners finished successfully")
	return nil
}

// Graph returns the dependency graph of services.
// This can be useful for visualization or analysis of the service dependencies.
func (c *Container) Graph() *dag.DAG[string, ServiceDef] {
	return c.graph
}

// addDependencyVertex adds a service to the dependency graph and recursively adds its dependencies.
// If parent is not nil, it also adds an edge from parent to service in the graph.
// This method is used during container initialization to build the complete dependency graph.
func (c *Container) addDependencyVertex(service ServiceDef, parent ServiceDef) error {
	if err := c.graph.AddVertexIfNotExist(service); err != nil {
		return err
	}

	if parent != nil {
		if err := c.graph.AddEdgeIfNotExist(parent.Name(), service.Name()); err != nil {
			return err
		}
	}
	m := service.Make()
	if isNil(m) {
		return nil
	}
	val := reflect.ValueOf(m)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if !val.IsValid() {
		return nil
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		dependencyName := typ.Field(i).Type.String()
		if childService, ok := c.services[dependencyName]; ok {
			if err := c.addDependencyVertex(childService, service); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Container) injectLoggerIntoField(field reflect.Value, target any) {
	logger := slog.Default()
	for _, attrSetter := range c.config.AttrSetters {
		name, value := attrSetter(target)
		logger = logger.With(name, value)
	}
	field.Set(reflect.ValueOf(logger))
}

// serviceHash returns a unique identifier for a service, which is its name.
// This is used as the vertex identifier in the dependency graph.
func serviceHash(service ServiceDef) string {
	return service.Name()
}

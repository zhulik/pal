package pal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strings"

	typetostring "github.com/samber/go-type-to-string"

	"golang.org/x/sync/errgroup"

	"github.com/zhulik/pal/pkg/dag"
)

// Container is responsible for storing services, instances and the dependency graph
type Container struct {
	config *Config

	services map[string]ServiceDef
	graph    *dag.DAG[string, ServiceDef]
	logger   *slog.Logger

	runners *RunnerGroup
}

// NewContainer creates a new Container instance
func NewContainer(config *Config, services ...ServiceDef) *Container {
	services = flattenServices(services)
	index := make(map[string]ServiceDef)

	for _, service := range services {
		index[service.Name()] = service
	}

	return &Container{
		config:   config,
		services: index,
		graph:    dag.New[string, ServiceDef](),
		logger:   slog.With("palComponent", "Container"),
		runners:  &RunnerGroup{},
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

	for _, service := range c.graph.ReverseTopologicalOrder() {
		if err := service.Init(ctx); err != nil {
			c.logger.Error("Failed to initialize container", "error", err)
			return err
		}
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

		typeName := typetostring.GetReflectType(fieldType)

		dependency, err := c.Invoke(ctx, typeName)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			if errors.Is(err, ErrServiceInvalidArgumentsCount) {
				return fmt.Errorf("%w: '%s': %w", ErrFactoryServiceDependency, typeName, err)
			}
			return err
		}

		field.Set(reflect.ValueOf(dependency))
	}
	return nil
}

func (c *Container) Shutdown(ctx context.Context) error {
	c.logger.Debug("Shutting down all runners")
	err := c.runners.Stop(ctx)

	if err != nil {
		c.logger.Error("Runners failed to stop", "error", err)
		return err
	}

	for _, service := range c.graph.TopologicalOrder() {
		err = service.Shutdown(ctx)
		if err != nil {
			c.logger.Error("Failed to shutdown service. Shutdown sequence is interrupted", "service", service.Name(), "error", err)
			return err
		}
	}

	c.logger.Debug("Container shut down successfully")
	return nil
}

func (c *Container) HealthCheck(ctx context.Context) error {
	var wg errgroup.Group

	c.logger.Debug("Healthchecking services")

	for _, service := range c.graph.TopologicalOrder() {
		wg.Go(func() error {
			// Do not check pal again, this leads to recursion
			if service.Name() == "*github.com/zhulik/pal.Pal" {
				return nil
			}

			err := service.HealthCheck(ctx)
			if err != nil {
				return err
			}

			return nil
		})
	}

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
	services := slices.Collect(maps.Values(c.services))
	ok, err := c.runners.Run(ctx, services)
	if err != nil {
		c.logger.Error("Starting runners failed", "error", err)
		return err
	}
	if !ok {
		c.logger.Debug("No main runners found, exiting")
		return nil
	}
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
	c.graph.AddVertexIfNotExist(service.Name(), service)

	if parent != nil {
		if err := c.graph.AddEdge(parent.Name(), service.Name()); err != nil {
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
		dependencyName := typetostring.GetReflectType(typ.Field(i).Type)
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

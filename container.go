package pal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"sync"

	"github.com/dominikbraun/graph"
	"golang.org/x/sync/errgroup"

	"github.com/zhulik/pal/pkg/dag"
)

// Container is responsible for storing services, instances and the dependency graph
type Container struct {
	services map[string]ServiceDef
	graph    *dag.DAG[string, ServiceDef]
	logger   *slog.Logger

	runnerTasks errgroup.Group

	cancelMu sync.RWMutex
	cancel   context.CancelFunc
}

// NewContainer creates a new Container instance
func NewContainer(services ...ServiceDef) *Container {
	index := make(map[string]ServiceDef)

	for _, service := range services {
		index[service.Name()] = service
	}

	return &Container{
		services: index,
		graph:    dag.New(serviceHash),
		logger:   slog.With("palComponent", "Container"),
	}
}

func (c *Container) Validate(ctx context.Context) error {
	var errs []error

	for _, service := range c.services {
		errs = append(errs, service.Validate(ctx))
	}

	return errors.Join(errs...)
}

func (c *Container) Init(ctx context.Context) error {
	c.logger.Info("Building dependency tree...")

	for _, service := range c.services {
		if err := c.addDependencyVertex(service, nil); err != nil {
			return err
		}
	}

	adjMap, err := c.graph.AdjacencyMap()
	if err != nil {
		return err
	}

	order, err := graph.TopologicalSort[string, ServiceDef](c.graph)
	if err != nil {
		return err
	}
	slices.Reverse(order)

	c.logger.Info("Dependency tree built", "tree", adjMap, "order", order)

	return c.graph.InReverseTopologicalOrder(func(service ServiceDef) error {
		if err := service.Init(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (c *Container) Invoke(ctx context.Context, name string) (any, error) {
	service, ok := c.services[name]
	if !ok {
		return nil, fmt.Errorf("%w: '%s', known services: %s", ErrServiceNotFound, name, c.services)
	}

	instance, err := service.Instance(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s': %w", ErrServiceInitFailed, name, err)
	}

	return instance, nil
}

func (c *Container) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutting down runners by cancelling their root context
	c.cancelMu.RLock()
	if c.cancel != nil {
		c.cancel()
	}
	c.cancelMu.RUnlock()

	// Await for runners to exit and safe possible error.
	errs = append(errs, c.runnerTasks.Wait())

	c.graph.InTopologicalOrder(func(service ServiceDef) error { // nolint:errcheck
		err := service.Shutdown(ctx)
		if err != nil {
			c.logger.Warn("Shut down with error", "service", service.Name(), "error", err)
			errs = append(errs, err)
			return nil
		}

		c.logger.Info("Shut down successfully", "service", service.Name())
		return nil
	})

	return errors.Join(errs...)
}

func (c *Container) HealthCheck(ctx context.Context) error {
	var wg errgroup.Group

	c.graph.ForEachVertex(func(service ServiceDef) error { // nolint:errcheck
		wg.Go(func() error {
			// Do not check pal again, this leads to recursion
			if service.Name() == "*pal.Pal" {
				return nil
			}

			err := service.HealthCheck(ctx)
			if err != nil {
				c.logger.Warn("Health check failed", "service", service.Name(), "error", err)
				return err
			}

			c.logger.Debug("Health check successful", "service", service.Name())
			return nil
		})

		return nil
	})

	return wg.Wait()
}

func (c *Container) Services() map[string]ServiceDef {
	return c.services
}

func (c *Container) StartRunners(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c.cancelMu.Lock()
	c.cancel = cancel
	c.cancelMu.Unlock()

	for name, service := range c.services {
		if !service.IsRunner() {
			continue
		}

		runner, err := service.Instance(ctx)
		if err != nil {
			return err
		}

		c.runnerTasks.Go(tryWrap(func() error {
			c.logger.Info("Running", "service", name)
			err := runner.(Runner).Run(ctx)
			if err != nil {
				c.logger.Warn("Runner exited with error, scheduling shutdown", "service", name, "error", err)
				FromContext(ctx).Shutdown(err)
				return err
			}

			c.logger.Info("Runner finished successfully", "service", name)
			return nil
		}))
	}

	c.logger.Info("Waiting for runners to finish")
	err := c.runnerTasks.Wait()
	if err != nil {
		c.logger.Warn("Runners finished with error", "error", err)
		return nil
	}
	c.logger.Info("All runners finished successfully")
	return nil
}

func (c *Container) Graph() *dag.DAG[string, ServiceDef] {
	return c.graph
}

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
		field := typ.Field(i)

		if field.Type.Kind() == reflect.Interface {
			dependencyName := field.Type.String()
			if childService, ok := c.services[dependencyName]; ok {
				if err := c.addDependencyVertex(childService, service); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func serviceHash(service ServiceDef) string {
	return service.Name()
}

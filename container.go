package pal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"sync"

	"github.com/dominikbraun/graph"
	"golang.org/x/sync/errgroup"

	"github.com/zhulik/pal/pkg/dag"
)

var (
	emptyLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
)

// Container is responsible for storing services, instances and the dependency graph
type Container struct {
	services map[string]ServiceImpl
	graph    *dag.DAG[string, ServiceImpl]
	logger   Logger

	runnerTasks errgroup.Group

	cancelMu sync.RWMutex
	cancel   context.CancelFunc
}

// NewContainer creates a new Container instance
func NewContainer(services ...ServiceImpl) *Container {
	index := make(map[string]ServiceImpl)

	for _, service := range services {
		index[service.Name()] = service
	}

	return &Container{
		services: index,
		graph:    dag.New(serviceHash),
		logger:   emptyLogger,
	}
}

func (c *Container) SetLogger(logger Logger) {
	c.logger = logger
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

	order, err := graph.TopologicalSort[string, ServiceImpl](c.graph)
	if err != nil {
		return err
	}
	slices.Reverse(order)

	c.logger.Info("Dependency tree built", "tree", adjMap, "order", order)

	return c.graph.InReverseTopologicalOrder(func(service ServiceImpl) error {
		if service.IsSingleton() {
			c.logger.Info("Initializing", "service", service.Name())

			if err := service.Initialize(ctx); err != nil {
				return err
			}

			c.logger.Info("Initialized", "service", service.Name())
		}

		return nil
	})
}

func (c *Container) Invoke(ctx context.Context, name string) (any, error) {
	service, err := c.graph.Vertex(name)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s', known services: %s. %w", ErrServiceNotFound, name, c.Services(), err)
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

	c.graph.InTopologicalOrder(func(service ServiceImpl) error { // nolint:errcheck
		if !service.IsSingleton() {
			return nil
		}

		instance, _ := service.Instance(ctx)

		if shutdowner, ok := instance.(Shutdowner); ok {
			c.logger.Info("Shutting down", "service", service.Name())

			err := shutdowner.Shutdown(ctx)
			if err != nil {
				c.logger.Warn("Shut down with error", "service", service.Name(), "error", err)
				errs = append(errs, err)
				return nil
			}

			c.logger.Info("Shut down successfully", "service", service.Name())
		}

		return nil
	})

	return errors.Join(errs...)
}

func (c *Container) HealthCheck(ctx context.Context) error {
	var wg errgroup.Group

	c.graph.ForEachVertex(func(service ServiceImpl) error { // nolint:errcheck
		if !service.IsSingleton() {
			return nil
		}

		wg.Go(func() error {
			instance, _ := service.Instance(ctx)

			if healthChecker, ok := instance.(HealthChecker); ok {
				c.logger.Debug("Health checking", "service", service.Name())

				err := healthChecker.HealthCheck(ctx)
				if err != nil {
					c.logger.Warn("Health check failed", "service", service.Name(), "error", err)
					return err
				}

				c.logger.Debug("Health check successful", "service", service.Name())
			}
			return nil
		})

		return nil
	})

	return wg.Wait()
}

func (c *Container) Services() []ServiceImpl {
	return c.graph.Vertices()
}

func (c *Container) StartRunners(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c.cancelMu.Lock()
	c.cancel = cancel
	c.cancelMu.Unlock()

	for _, service := range c.Services() {
		if !service.IsRunner() {
			continue
		}

		runner, err := service.Instance(ctx)
		if err != nil {
			return err
		}

		c.runnerTasks.Go(func() error {
			c.logger.Info("Running", "service", service.Name())
			err := runner.(Runner).Run(ctx)
			if err != nil {
				c.logger.Warn("Runner exited with error, scheduling shutdown", "service", service.Name(), "error", err)
				FromContext(ctx).Shutdown(err)
				return err
			}

			c.logger.Info("Runner finished successfully", "service", service.Name())
			return nil
		})
	}

	c.logger.Info("Waiting for runners to finish")
	err := c.runnerTasks.Wait()
	if err != nil {
		c.logger.Warn("Runners finished with", "error", err)
	}
	c.logger.Info("All runners finished successfully")
	return err
}

func (c *Container) addDependencyVertex(service ServiceImpl, parent ServiceImpl) error {
	if err := c.graph.AddVertexIfNotExist(service); err != nil {
		return err
	}

	if parent != nil {
		if err := c.graph.AddEdgeIfNotExist(parent.Name(), service.Name()); err != nil {
			return err
		}
	}

	val := reflect.ValueOf(service.Make())
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

func serviceHash(service ServiceImpl) string {
	return service.Name()
}

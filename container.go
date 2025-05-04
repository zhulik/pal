package pal

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/sync/errgroup"

	"github.com/zhulik/pal/pkg/dag"
)

// Container is responsible for storing services, instances and the dependency graph
type Container struct {
	services map[string]ServiceImpl
	graph    *dag.DAG[string, ServiceImpl]
	log      LoggerFn

	runnerTasks errgroup.Group
}

// NewContainer creates a new Container instance
func NewContainer(services ...ServiceImpl) *Container {
	index := make(map[string]ServiceImpl)

	for _, service := range services {
		index[service.Name()] = service
	}

	return &Container{
		services: index,
		log:      func(string, ...any) {},
		graph:    dag.New(serviceHash),
	}
}

func (c *Container) SetLogger(log LoggerFn) {
	c.log = log
}

func (c *Container) Validate(ctx context.Context) error {
	var errs []error

	for _, service := range c.services {
		errs = append(errs, service.Validate(ctx))
	}

	return errors.Join(errs...)
}

func (c *Container) Init(ctx context.Context) error {
	for _, service := range c.services {
		if err := c.addDependencyVertex(service, nil); err != nil {
			return err
		}
	}

	// file, _ := os.Initialize("./mygraph.gv")
	// _ = draw.DOT(c.graph, file)

	err := c.graph.InReverseTopologicalOrder(func(service ServiceImpl) error {
		if service.IsSingleton() {
			c.log("initializing %s", service.Name())

			if err := service.Initialize(ctx); err != nil {
				return err
			}

			c.log("%s initialized", service.Name())
		}

		return nil
	})

	return err
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

	c.graph.InTopologicalOrder(func(service ServiceImpl) error { // nolint:errcheck
		if !service.IsSingleton() {
			return nil
		}

		// In topological order runners appear first naturally, once there are no more runners in the list,
		// we can wait for them to finish.
		if !service.IsRunner() {
			errs = append(errs, c.runnerTasks.Wait())
		}

		instance, _ := service.Instance(ctx)

		if shutdowner, ok := instance.(Shutdowner); ok {
			c.log("shutting down %s", service.Name())

			err := shutdowner.Shutdown(ctx)
			if err != nil {
				c.log("%s shut down with error=%+v", service.Name(), err)
				errs = append(errs, err)
				return nil
			}

			c.log("%s shut down successfully", service.Name())
		}

		return nil
	})

	return errors.Join(errs...)
}

func (c *Container) HealthCheck(ctx context.Context) error {
	return c.graph.ForEachVertex(func(service ServiceImpl) error { // nolint:errcheck
		if service.IsSingleton() {
			instance, _ := service.Instance(ctx)

			if healthChecker, ok := instance.(HealthChecker); ok {
				c.log("health checking %s", service.Name())

				err := healthChecker.HealthCheck(ctx)
				if err != nil {
					c.log("%s failed health check error=%+v", service.Name(), err)
					return err
				}

				c.log("%s passed health check successfully", service.Name())
			}
		}

		return nil
	})
}

func (c *Container) Services() []ServiceImpl {
	return c.graph.Vertices()
}

func (c *Container) runners(ctx context.Context) map[string]Runner {
	runners := map[string]Runner{}

	c.graph.ForEachVertex(func(service ServiceImpl) error { // nolint:errcheck
		if service.IsRunner() {
			if runner, err := service.Instance(ctx); err == nil {
				runners[service.Name()] = runner.(Runner)
			}
		}
		return nil
	})

	return runners
}

func (c *Container) StartRunners(ctx context.Context) error {
	for name, runner := range c.runners(ctx) {
		c.runnerTasks.Go(func() error {
			c.log("running %s", name)
			err := runner.Run(ctx)
			if err != nil {
				c.log("runner %s exited with error='%+v'", name, err)
				return err
			}

			c.log("runner %s finished successfully", name)
			return nil
		})
	}
	c.log("waiting for runners to finish")
	err := c.runnerTasks.Wait()
	if err != nil {
		c.log("all runners finished with error='%+v'", err)
	}
	c.log("all runners finished successfully")
	return err
}

func (c *Container) addDependencyVertex(service ServiceImpl, parent ServiceImpl) error {
	if _, err := c.graph.Vertex(service.Name()); err == nil {
		// service and all it's dependencies already exist in the graph.
		return nil
	}

	if err := c.graph.AddVertexIfNotExist(service); err != nil {
		return err
	}

	if parent != nil {
		if err := c.graph.AddEdgeIfNotExist(parent.Name(), service.Name()); err != nil {
			return err
		}
	}

	instance := service.Make()

	val := reflect.ValueOf(instance)
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

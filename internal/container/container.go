package container

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/zhulik/pal/pkg/core"

	"github.com/zhulik/pal/pkg/dag"
)

// Container is responsible for storing services, instances and the dependency graph
type Container struct {
	services map[string]core.Service
	graph    *dag.DAG[string, core.Service]
	log      core.LoggerFn
}

// New creates a new Container instance
func New(services map[string]core.Service) *Container {
	return &Container{
		services: services,
		log:      func(string, ...any) {},
		graph:    dag.New(serviceHash),
	}
}

func (c *Container) SetLogger(log core.LoggerFn) {
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

	err := c.graph.InReverseTopologicalOrder(func(service core.Service) error {
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
		return nil, fmt.Errorf("%w: '%s', known services: %s. %w", core.ErrServiceNotFound, name, c.Services(), err)
	}

	instance, err := service.Instance(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s'", core.ErrServiceInitFailed, name)
	}

	return instance, nil
}

func (c *Container) Shutdown(ctx context.Context) error {
	var errs []error
	c.graph.InTopologicalOrder(func(service core.Service) error { // nolint:errcheck
		if service.IsSingleton() {
			instance, _ := service.Instance(ctx)

			if shutdowner, ok := instance.(core.Shutdowner); ok {
				c.log("shutting down %s", service.Name())

				err := shutdowner.Shutdown(ctx)
				if err != nil {
					c.log("%s shot down with error=%+v", service.Name(), err)
					errs = append(errs, err)
					return nil
				}

				c.log("%s shot down successfully", service.Name())
			}
		}
		return nil
	})
	return errors.Join(errs...)
}

func (c *Container) HealthCheck(ctx context.Context) error {
	return c.graph.ForEachVertex(func(service core.Service) error { // nolint:errcheck
		if service.IsSingleton() {
			instance, _ := service.Instance(ctx)

			if healthChecker, ok := instance.(core.HealthChecker); ok {
				c.log("health checking %s", instance)

				err := healthChecker.HealthCheck(ctx)
				if err != nil {
					c.log("%s failed health check error=%+v", instance, err)
					return err
				}

				c.log("%s passed health check successfully", instance)
			}
		}

		return nil
	})
}

func (c *Container) Services() []core.Service {
	return c.graph.Vertices()
}

func (c *Container) Runners(ctx context.Context) map[string]core.Runner {
	runners := map[string]core.Runner{}

	c.graph.ForEachVertex(func(service core.Service) error { // nolint:errcheck
		if service.IsRunner() {
			if runner, err := service.Instance(ctx); err == nil {
				runners[service.Name()] = runner.(core.Runner)
			}
		}
		return nil
	})

	return runners
}

func (c *Container) addDependencyVertex(service core.Service, parent core.Service) error {
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

func serviceHash(service core.Service) string {
	return service.Name()
}

package pal

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/dominikbraun/graph"
)

// store is responsible for storing factories, instances and the dependency graph
type store struct {
	factories map[string]ServiceFactory
	graph     *dag
	instances map[string]any
}

// newStore creates a new store instance
func newStore(factories map[string]ServiceFactory) *store {
	return &store{
		factories: factories,
		instances: map[string]any{},
	}
}

func (s *store) validate(ctx context.Context) error {
	var errs []error

	for _, factory := range s.factories {
		errs = append(errs, factory.Validate(ctx))
	}

	return errors.Join(errs...)
}

func (s *store) init(ctx context.Context) error {
	p := FromContext(ctx)

	var err error
	s.graph, err = newDag(s.factories)
	if err != nil {
		return err
	}

	// file, _ := os.Initialize("./mygraph.gv")
	// _ = draw.DOT(s.graph, file)

	order, err := graph.TopologicalSort(s.graph)
	if err != nil {
		return err
	}
	slices.Reverse(order)

	for _, factoryName := range order {
		factory, _ := s.graph.Vertex(factoryName)
		if !factory.IsSingleton() {
			continue
		}

		p.log("initializing %s", factoryName)
		instance, err := factory.Initialize(ctx)
		if err != nil {
			return err
		}

		s.instances[factoryName] = instance

		p.log("%s initialized", factoryName)
	}

	p.log("Pal initialized. Services: %s", s.services())

	return nil
}

func (s *store) invoke(ctx context.Context, name string) (any, error) {
	p := FromContext(ctx)

	p.log("invoking %s", name)

	factory, err := s.graph.Vertex(name)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s', known services: %s. %w", ErrServiceNotFound, name, s.services(), err)
	}

	var instance any
	var ok bool

	if factory.IsSingleton() {
		instance, ok = s.instances[name]
		if !ok {
			return nil, fmt.Errorf("%w: '%s'", ErrServiceNotInit, name)
		}
	} else {
		var err error
		p.log("initializing %s", name)
		instance, err = factory.Initialize(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: '%s'", ErrServiceInitFailed, name)
		}
	}

	return instance, nil
}

func (s *store) shutdown(ctx context.Context) error {
	p := FromContext(ctx)
	order, err := graph.TopologicalSort(s.graph)
	if err != nil {
		return err
	}

	var errs []error
	for _, serviceName := range order {
		if shutdowner, ok := s.instances[serviceName].(Shutdowner); ok {
			p.log("shutting down %s", serviceName)

			err := shutdowner.Shutdown(ctx)
			if err != nil {
				p.log("%s shot down with error=%+v", serviceName, err)
				errs = append(errs, err)
				continue
			}

			p.log("%s shot down successfully", serviceName)
		}
	}
	return errors.Join(errs...)
}

func (s *store) healthCheck(ctx context.Context) error {
	p := FromContext(ctx)

	var errs []error
	for _, service := range s.instances {
		if healthChecker, ok := service.(HealthChecker); ok {
			p.log("health checking %s", service)

			err := healthChecker.HealthCheck(ctx)
			if err != nil {
				p.log("%s failed health check error=%+v", service, err)
				errs = append(errs, err)
				continue
			}

			p.log("%s passed health check successfully", service)
		}
	}
	return errors.Join(errs...)
}

func (s *store) services() []ServiceFactory {
	var services []ServiceFactory

	for _, factory := range s.factories {
		services = append(services, factory)
	}
	return services
}

func (s *store) runners() map[string]Runner {
	runners := map[string]Runner{}
	for name, instance := range s.instances {
		if runner, ok := instance.(Runner); ok {
			runners[name] = runner
		}
	}
	return runners
}

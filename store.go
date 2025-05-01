package pal

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/dominikbraun/graph"
)

// store is responsible for storing factories, instances and the dependency graph
type store struct {
	factories map[string]ServiceFactory
	graph     graph.Graph[string, ServiceFactory]
	instances map[string]any
}

func serviceFactoryHash(factory ServiceFactory) string {
	return factory.Name()
}

// newStore creates a new store instance
func newStore(factories map[string]ServiceFactory) *store {
	return &store{
		factories: factories,
		instances: map[string]any{},
		graph:     graph.New(serviceFactoryHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}
}

func (s *store) init(ctx context.Context) error {
	p := FromContext(ctx)

	err := s.buildDAG()
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

func (s *store) buildDAG() error {
	for _, factory := range s.factories {
		if err := s.addDependencyVertex(factory, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *store) addDependencyVertex(factory ServiceFactory, parent ServiceFactory) error {
	if err := s.graph.AddVertex(factory); err != nil {
		if !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return err
		}
	}

	if parent != nil {
		if err := s.graph.AddEdge(parent.Name(), factory.Name()); err != nil {
			if !errors.Is(err, graph.ErrEdgeAlreadyExists) {
				return err
			}
		}
	}

	instance := factory.Make()

	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Type.Kind() == reflect.Interface {
			dependencyName := field.Type.String()
			if childFactory, ok := s.factories[dependencyName]; ok {
				if err := s.addDependencyVertex(childFactory, factory); err != nil {
					return err
				}
			}
		}
	}

	return nil
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

func (s *store) validate(ctx context.Context) error {
	var errs []error

	for _, factory := range s.factories {
		errs = append(errs, factory.Validate(ctx))
	}

	return errors.Join(errs...)
}

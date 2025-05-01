package pal

import (
	"context"
	"errors"
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
		factory := s.factories[factoryName]
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

	p.log("Pal initialized. Services: %s, runners: %s", s.services(), s.runners())

	return nil
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
	runners := s.runners()

	for _, runner := range runners {
		err := s.addDependencyVertex(runner, nil)
		if err != nil {
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
			return err
		}
	}

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
			if childFactory, ok := s.factories[dependencyName]; ok {
				if err := s.addDependencyVertex(childFactory, factory); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *store) services() []string {
	var names []string
	for name := range s.factories {
		names = append(names, name)
	}
	return names
}

func (s *store) runners() []ServiceFactory {
	var runners []ServiceFactory
	for _, factory := range s.factories {
		if factory.IsRunner() {
			runners = append(runners, factory)
		}
	}
	return runners
}

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
	graph     graph.Graph[string, string]
	instances map[string]any
}

// newStore creates a new store instance
func newStore(factories map[string]ServiceFactory) *store {
	return &store{
		factories: factories,
		instances: map[string]any{},
		graph:     graph.New(graph.StringHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}
}

func (s *store) init(ctx context.Context, p *Pal) error {
	defer func() {
		// TODO: if init fails with an error - try to gracefully shutdown already initialized dependencies.
	}()

	// Initialize dependencies staring from the leaves. The more dependants an services, the earlier it will be initialized.

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

		instance, err := factory.Initialize(ctx)
		if err != nil {
			return err
		}

		s.instances[factoryName] = instance
	}

	p.log("Initialized. Services: %s, runners: %s", p.Services(), p.Runners())

	for name, instance := range s.instances {
		if runner, ok := instance.(Runner); ok {
			// TODO: make it possible to wait for all runners.
			go func() {
				// TODO: use a custom context struct?
				p.log("running %s", name)
				ctx := context.WithValue(context.Background(), CtxValue, s)
				err := runner.Run(ctx)

				if err != nil {
					p.Error(err)
				}
			}()
		}
	}
	return nil
}

func (s *store) shutdown(ctx context.Context) error {
	order, err := graph.TopologicalSort(s.graph)
	if err != nil {
		return err
	}

	var errs []error
	for _, serviceName := range order {
		if shutdowner, ok := s.instances[serviceName].(Shutdowner); ok {
			errs = append(errs, shutdowner.Shutdown(ctx))
		}
	}
	return errors.Join(errs...)
}

func (s *store) healthCheck(ctx context.Context) error {
	var errs []error
	for _, service := range s.instances {
		if healthChecker, ok := service.(HealthChecker); ok {
			errs = append(errs, healthChecker.HealthCheck(ctx))
		}
	}
	return errors.Join(errs...)
}

func (s *store) buildDAG() error {
	runners := s.runners()

	for _, runner := range runners {
		err := s.addDependencyVertex(runner, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *store) addDependencyVertex(name string, parent string) error {
	if err := s.graph.AddVertex(name); err != nil {
		if !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return err
		}
	}

	if parent != "" {
		if err := s.graph.AddEdge(parent, name); err != nil {
			return err
		}
	}

	factory := s.factories[name]

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
			if _, ok := s.factories[dependencyName]; ok {
				if err := s.addDependencyVertex(dependencyName, name); err != nil {
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

func (s *store) runners() []string {
	var runners []string
	for name, factory := range s.factories {
		if factory.IsRunner() {
			runners = append(runners, name)
		}
	}
	return runners
}

package pal

import (
	"errors"
	"reflect"

	"github.com/dominikbraun/graph"
)

type dag struct {
	graph.Graph[string, ServiceFactory]
}

func newDag(factories map[string]ServiceFactory) (*dag, error) {
	d := &dag{
		graph.New(serviceFactoryHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}

	for _, factory := range factories {
		if err := d.addDependencyVertex(factory, nil, factories); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (d *dag) addDependencyVertex(factory ServiceFactory, parent ServiceFactory, factories map[string]ServiceFactory) error {
	if err := d.AddVertex(factory); err != nil {
		if !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return err
		}
	}

	if parent != nil {
		if err := d.AddEdge(parent.Name(), factory.Name()); err != nil {
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
			if childFactory, ok := factories[dependencyName]; ok {
				if err := d.addDependencyVertex(childFactory, factory, factories); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func serviceFactoryHash(factory ServiceFactory) string {
	return factory.Name()
}

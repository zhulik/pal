package pal

import (
	"errors"
	"reflect"
	"slices"

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

func (d *dag) Vertices() []ServiceFactory {
	// graph.Graph does not have a way to get the list of its vertices.
	// https://github.com/dominikbraun/graph/pull/149

	adjMap, _ := d.AdjacencyMap()

	vertices := make([]ServiceFactory, 0, len(adjMap))
	for hash := range adjMap {
		vertex, _ := d.Vertex(hash)

		vertices = append(vertices, vertex)
	}

	return vertices
}

func (d *dag) ForEachVertex(fn func(ServiceFactory) error) error {
	for _, factory := range d.Vertices() {
		if err := fn(factory); err != nil {
			return err
		}
	}
	return nil
}

func (d *dag) InReverseTopologicalOrder(fn func(ServiceFactory) error) error {
	order, err := graph.TopologicalSort(d.Graph)
	if err != nil {
		return err
	}
	slices.Reverse(order)

	for _, hash := range order {
		factory, _ := d.Vertex(hash)

		if err := fn(factory); err != nil {
			return err
		}
	}
	return nil
}

func (d *dag) InTopologicalOrder(fn func(ServiceFactory) error) error {
	order, err := graph.TopologicalSort(d.Graph)
	if err != nil {
		return err
	}

	for _, hash := range order {
		factory, _ := d.Vertex(hash)

		if err := fn(factory); err != nil {
			return err
		}
	}
	return nil
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

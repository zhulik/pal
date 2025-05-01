package pal

import (
	"errors"
	"reflect"
	"slices"

	"github.com/dominikbraun/graph"
)

type dag struct {
	graph.Graph[string, Service]
}

func newDag(services map[string]Service) (*dag, error) {
	d := &dag{
		graph.New(serviceHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}

	for _, service := range services {
		if err := d.addDependencyVertex(service, nil, services); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (d *dag) Vertices() []Service {
	// graph.Graph does not have a way to get the list of its vertices.
	// https://github.com/dominikbraun/graph/pull/149

	adjMap, _ := d.AdjacencyMap()

	vertices := make([]Service, 0, len(adjMap))
	for hash := range adjMap {
		vertex, _ := d.Vertex(hash)

		vertices = append(vertices, vertex)
	}

	return vertices
}

func (d *dag) ForEachVertex(fn func(Service) error) error {
	for _, service := range d.Vertices() {
		if err := fn(service); err != nil {
			return err
		}
	}
	return nil
}

func (d *dag) InReverseTopologicalOrder(fn func(Service) error) error {
	order, err := graph.TopologicalSort(d.Graph)
	if err != nil {
		return err
	}
	slices.Reverse(order)

	for _, hash := range order {
		service, _ := d.Vertex(hash)

		if err := fn(service); err != nil {
			return err
		}
	}
	return nil
}

func (d *dag) InTopologicalOrder(fn func(Service) error) error {
	order, err := graph.TopologicalSort(d.Graph)
	if err != nil {
		return err
	}

	for _, hash := range order {
		service, _ := d.Vertex(hash)

		if err := fn(service); err != nil {
			return err
		}
	}
	return nil
}

func (d *dag) addDependencyVertex(service Service, parent Service, services map[string]Service) error {
	if err := d.AddVertex(service); err != nil {
		if !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return err
		}
	}

	if parent != nil {
		if err := d.AddEdge(parent.Name(), service.Name()); err != nil {
			if !errors.Is(err, graph.ErrEdgeAlreadyExists) {
				return err
			}
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
			if childService, ok := services[dependencyName]; ok {
				if err := d.addDependencyVertex(childService, service, services); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func serviceHash(service Service) string {
	return service.Name()
}

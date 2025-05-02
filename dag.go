package pal

import (
	"errors"
	"slices"

	"github.com/dominikbraun/graph"
)

type dag struct {
	graph.Graph[string, Service]
}

func newDag() *dag {
	return &dag{
		graph.New(serviceHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}
}

func (d *dag) AddVertexIfNotExist(service Service) error {
	err := d.AddVertex(service)
	if errors.Is(err, graph.ErrVertexAlreadyExists) {
		return nil
	}
	return err
}

func (d *dag) AddEdgeIfNotExist(sourceHash, targetHash string, options ...func(*graph.EdgeProperties)) error {
	err := d.AddEdge(sourceHash, targetHash, options...)
	if errors.Is(err, graph.ErrEdgeAlreadyExists) {
		return nil
	}
	return err
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

func serviceHash(service Service) string {
	return service.Name()
}

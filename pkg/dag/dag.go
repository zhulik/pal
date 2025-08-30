package dag

import (
	"cmp"
	"errors"
	"slices"

	"github.com/dominikbraun/graph"
)

type DAG[K cmp.Ordered, T any] struct {
	graph.Graph[K, T]
}

func New[K cmp.Ordered, T any](hash graph.Hash[K, T]) *DAG[K, T] {
	return &DAG[K, T]{
		graph.New(hash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}
}

func (d *DAG[K, T]) AddVertexIfNotExist(v T) error {
	err := d.AddVertex(v)
	if errors.Is(err, graph.ErrVertexAlreadyExists) {
		return nil
	}
	return err
}

func (d *DAG[K, T]) AddEdgeIfNotExist(sourceHash, targetHash K, options ...func(*graph.EdgeProperties)) error {
	err := d.AddEdge(sourceHash, targetHash, options...)
	if errors.Is(err, graph.ErrEdgeAlreadyExists) {
		return nil
	}
	return err
}

func (d *DAG[K, T]) Vertices() []T {
	// graph.Graph does not have a way to get the list of its vertices.
	// https://github.com/dominikbraun/graph/pull/149

	adjMap, _ := d.AdjacencyMap()

	keys := make([]K, 0, len(adjMap))
	for hash := range adjMap {
		keys = append(keys, hash)
	}

	slices.Sort(keys)

	vertices := make([]T, 0, len(adjMap))
	for _, hash := range keys {
		vertex, _ := d.Vertex(hash)
		vertices = append(vertices, vertex)
	}

	return vertices
}

func (d *DAG[K, T]) ForEachVertex(fn func(T) error) error {
	for _, v := range d.Vertices() {
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DAG[K, T]) TopologicalOrder() ([]K, error) {
	return graph.TopologicalSort(d.Graph)
}

func (d *DAG[K, T]) ReverseTopologicalOrder() ([]K, error) {
	order, err := d.TopologicalOrder()
	if err != nil {
		return nil, err
	}
	slices.Reverse(order)
	return order, nil
}

func (d *DAG[K, T]) InReverseTopologicalOrder(fn func(T) error) error {
	order, err := d.ReverseTopologicalOrder()
	if err != nil {
		return err
	}

	for _, hash := range order {
		v, _ := d.Vertex(hash)

		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DAG[K, T]) InTopologicalOrder(fn func(T) error) error {
	order, err := graph.TopologicalSort(d.Graph)
	if err != nil {
		return err
	}

	for _, hash := range order {
		v, _ := d.Vertex(hash)

		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

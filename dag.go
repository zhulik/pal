package pal

import (
	"errors"
	"slices"

	"github.com/dominikbraun/graph"
)

type dag[K comparable, T any] struct {
	graph.Graph[K, T]
}

func newDag[K comparable, T any](hash graph.Hash[K, T]) *dag[K, T] {
	return &dag[K, T]{
		graph.New(hash, graph.Directed(), graph.Acyclic(), graph.PreventCycles()),
	}
}

func (d *dag[K, T]) AddVertexIfNotExist(v T) error {
	err := d.AddVertex(v)
	if errors.Is(err, graph.ErrVertexAlreadyExists) {
		return nil
	}
	return err
}

func (d *dag[K, T]) AddEdgeIfNotExist(sourceHash, targetHash K, options ...func(*graph.EdgeProperties)) error {
	err := d.AddEdge(sourceHash, targetHash, options...)
	if errors.Is(err, graph.ErrEdgeAlreadyExists) {
		return nil
	}
	return err
}

func (d *dag[K, T]) Vertices() []T {
	// graph.Graph does not have a way to get the list of its vertices.
	// https://github.com/dominikbraun/graph/pull/149

	adjMap, _ := d.AdjacencyMap()

	vertices := make([]T, 0, len(adjMap))
	for hash := range adjMap {
		vertex, _ := d.Vertex(hash)

		vertices = append(vertices, vertex)
	}

	return vertices
}

func (d *dag[K, T]) ForEachVertex(fn func(T) error) error {
	for _, v := range d.Vertices() {
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *dag[K, T]) InReverseTopologicalOrder(fn func(T) error) error {
	order, err := graph.TopologicalSort(d.Graph)
	if err != nil {
		return err
	}
	slices.Reverse(order)

	for _, hash := range order {
		v, _ := d.Vertex(hash)

		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *dag[K, T]) InTopologicalOrder(fn func(T) error) error {
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

package dag

import (
	"errors"
	"iter"
	"maps"
	"slices"
)

var (
	ErrEdgeAlreadyExists = errors.New("edge already exists")
	ErrCycleDetected     = errors.New("cycle detected")
	ErrVertexNotFound    = errors.New("vertex not found")
)

type DAG[ID comparable, T any] struct {
	vertices map[ID]T
	edges    map[ID]map[ID]bool // adjacency list: source -> set of targets
	inDegree map[ID]int         // in-degree count for each vertex
}

func New[ID comparable, T any]() *DAG[ID, T] {
	return &DAG[ID, T]{
		vertices: make(map[ID]T),
		edges:    make(map[ID]map[ID]bool),
		inDegree: make(map[ID]int),
	}
}

// VertexCount returns the total number of vertices in the DAG
func (d *DAG[ID, T]) VertexCount() int {
	return len(d.vertices)
}

// EdgeCount returns the total number of edges in the DAG
func (d *DAG[ID, T]) EdgeCount() int {
	count := 0
	for _, targets := range d.edges {
		count += len(targets)
	}
	return count
}

// VertexExists checks if a vertex with the given ID exists
func (d *DAG[ID, T]) VertexExists(id ID) bool {
	_, exists := d.vertices[id]
	return exists
}

// EdgeExists checks if an edge from source to target exists
func (d *DAG[ID, T]) EdgeExists(source, target ID) bool {
	if !d.VertexExists(source) {
		return false
	}
	return d.edges[source][target]
}

// GetVertex returns the vertex data for the given ID and whether it exists
func (d *DAG[ID, T]) GetVertex(id ID) (T, bool) {
	val, exists := d.vertices[id]
	return val, exists
}

// GetInDegree returns the in-degree (number of incoming edges) for a vertex
func (d *DAG[ID, T]) GetInDegree(id ID) int {
	return d.inDegree[id]
}

// GetOutDegree returns the out-degree (number of outgoing edges) for a vertex
func (d *DAG[ID, T]) GetOutDegree(id ID) int {
	if !d.VertexExists(id) {
		return 0
	}
	return len(d.edges[id])
}

func (d *DAG[ID, T]) AddVertexIfNotExist(id ID, v T) {
	if _, exists := d.vertices[id]; !exists {
		d.vertices[id] = v
		d.edges[id] = make(map[ID]bool)
		d.inDegree[id] = 0
	}
}

func (d *DAG[ID, T]) AddEdge(source, target ID) error {
	// Check if both vertices exist
	if !d.VertexExists(source) {
		return ErrVertexNotFound
	}
	if !d.VertexExists(target) {
		return ErrVertexNotFound
	}

	// Check if edge already exists
	if d.edges[source][target] {
		return ErrEdgeAlreadyExists
	}

	// Add the edge
	d.edges[source][target] = true
	d.inDegree[target]++

	// Check for cycles using DFS
	if d.hasCycle() {
		// Remove the edge if it creates a cycle
		d.edges[source][target] = false
		d.inDegree[target]--
		return ErrCycleDetected
	}

	return nil
}

func (d *DAG[ID, T]) TopologicalOrder() iter.Seq2[ID, T] {
	// Create a copy of in-degree counts
	inDegreeCopy := make(map[ID]int)
	maps.Copy(inDegreeCopy, d.inDegree)

	// Find all vertices with in-degree 0
	var queue []ID
	for id, count := range inDegreeCopy {
		if count == 0 {
			queue = append(queue, id)
		}
	}

	// Kahn's algorithm for topological sorting
	return func(yield func(ID, T) bool) {
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if !yield(current, d.vertices[current]) {
				break
			}

			// Reduce in-degree of all neighbors
			for neighbor := range d.edges[current] {
				inDegreeCopy[neighbor]--
				if inDegreeCopy[neighbor] == 0 {
					queue = append(queue, neighbor)
				}
			}
		}
	}
}

func (d *DAG[ID, T]) ReverseTopologicalOrder() iter.Seq2[ID, T] {
	var result []ID
	for id := range d.TopologicalOrder() {
		result = append(result, id)
	}
	slices.Reverse(result)

	return func(yield func(ID, T) bool) {
		for _, id := range result {
			if !yield(id, d.vertices[id]) {
				break
			}
		}
	}
}

// Helper method to detect cycles using DFS
func (d *DAG[ID, T]) hasCycle() bool {
	visited := make(map[ID]bool)
	recStack := make(map[ID]bool)

	var dfs func(ID) bool
	dfs = func(vertex ID) bool {
		if recStack[vertex] {
			return true // Back edge found, cycle detected
		}
		if visited[vertex] {
			return false // Already processed
		}

		visited[vertex] = true
		recStack[vertex] = true

		for neighbor := range d.edges[vertex] {
			if dfs(neighbor) {
				return true
			}
		}

		recStack[vertex] = false
		return false
	}

	for vertex := range d.vertices {
		if !visited[vertex] {
			if dfs(vertex) {
				return true
			}
		}
	}

	return false
}

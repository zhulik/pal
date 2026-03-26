package dag

import (
	"cmp"
	"errors"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
)

var (
	ErrEdgeAlreadyExists = errors.New("edge already exists")
	ErrCycleDetected     = errors.New("cycle detected")
	ErrVertexNotFound    = errors.New("vertex not found")
)

type CycleError[ID cmp.Ordered] struct {
	Cycle []ID
}

func (e *CycleError[ID]) Error() string {
	if len(e.Cycle) == 0 {
		return ErrCycleDetected.Error()
	}

	parts := make([]string, len(e.Cycle))
	for i, vertex := range e.Cycle {
		parts[i] = fmt.Sprint(vertex)
	}

	return fmt.Sprintf("%s: %s", ErrCycleDetected, strings.Join(parts, " -> "))
}

func (e *CycleError[ID]) Unwrap() error {
	return ErrCycleDetected
}

type DAG[ID cmp.Ordered, T any] struct {
	vertices map[ID]T
	edges    map[ID]map[ID]bool // adjacency list: source -> set of targets
	inDegree map[ID]int         // in-degree count for each vertex
}

func New[ID cmp.Ordered, T any]() *DAG[ID, T] {
	return &DAG[ID, T]{
		vertices: make(map[ID]T),
		edges:    make(map[ID]map[ID]bool),
		inDegree: make(map[ID]int),
	}
}

// Vertices returns the vertices of the DAG
func (d *DAG[ID, T]) Vertices() map[ID]T {
	return d.vertices
}

// Edges returns the edges of the DAG
func (d *DAG[ID, T]) Edges() map[ID]map[ID]bool {
	return d.edges
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
	if cycle, hasCycle := d.hasCycle(); hasCycle {
		// Remove the edge if it creates a cycle
		delete(d.edges[source], target)
		d.inDegree[target]--
		return &CycleError[ID]{Cycle: cycle}
	}

	return nil
}

func (d *DAG[ID, T]) AddEdgeIfNotExist(source, target ID) error {
	if d.EdgeExists(source, target) {
		return nil
	}
	return d.AddEdge(source, target)
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
func (d *DAG[ID, T]) hasCycle() ([]ID, bool) {
	visited := make(map[ID]bool)
	recStack := make(map[ID]bool)
	path := make([]ID, 0, len(d.vertices))

	var cycle []ID

	var dfs func(ID) bool
	dfs = func(vertex ID) bool {
		visited[vertex] = true
		recStack[vertex] = true
		path = append(path, vertex)

		neighbors := make([]ID, 0, len(d.edges[vertex]))
		for neighbor := range d.edges[vertex] {
			neighbors = append(neighbors, neighbor)
		}
		slices.Sort(neighbors)

		for _, neighbor := range neighbors {
			if recStack[neighbor] {
				start := slices.Index(path, neighbor)
				cycle = append([]ID{}, path[start:]...)
				cycle = append(cycle, neighbor)
				return true
			}

			if visited[neighbor] {
				continue
			}

			if dfs(neighbor) {
				return true
			}
		}

		path = path[:len(path)-1]
		recStack[vertex] = false
		return false
	}

	vertices := make([]ID, 0, len(d.vertices))
	for vertex := range d.vertices {
		vertices = append(vertices, vertex)
	}
	slices.Sort(vertices)

	for _, vertex := range vertices {
		if !visited[vertex] {
			if dfs(vertex) {
				return cycle, true
			}
		}
	}

	return nil, false
}

package dag

import (
	"errors"
	"testing"

	"github.com/dominikbraun/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hashFn(i int) int { return i }

// TestNew tests the New function
func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("creates a new DAG with hash function", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		assert.NotNil(t, dag)
		assert.NotNil(t, dag.Graph)
	})
}

// TestAddVertexIfNotExist tests the AddVertexIfNotExist method
func TestAddVertexIfNotExist(t *testing.T) {
	t.Parallel()

	t.Run("adds a vertex if it doesn't exist", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		err := dag.AddVertexIfNotExist(1)
		assert.NoError(t, err)

		// Verify vertex was added
		vertex, err := dag.Vertex(1)
		assert.NoError(t, err)
		assert.Equal(t, 1, vertex)
	})

	t.Run("does nothing if vertex already exists", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Add vertex first time
		err := dag.AddVertex(1)
		require.NoError(t, err)

		// Try to add same vertex again
		err = dag.AddVertexIfNotExist(1)
		assert.NoError(t, err)

		// Verify vertex still exists
		vertex, err := dag.Vertex(1)
		assert.NoError(t, err)
		assert.Equal(t, 1, vertex)
	})
}

// TestAddEdgeIfNotExist tests the AddEdgeIfNotExist method
func TestAddEdgeIfNotExist(t *testing.T) {
	t.Parallel()

	t.Run("adds an edge if it doesn't exist", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Add vertices first
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))

		// Add edge
		err := dag.AddEdgeIfNotExist(1, 2)
		assert.NoError(t, err)

		// Verify edge was added
		adjMap, err := dag.AdjacencyMap()
		assert.NoError(t, err)
		assert.Contains(t, adjMap[1], 2)
	})

	t.Run("does nothing if edge already exists", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Add vertices first
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))

		// Add edge first time
		err := dag.AddEdge(1, 2)
		require.NoError(t, err)

		// Try to add same edge again
		err = dag.AddEdgeIfNotExist(1, 2)
		assert.NoError(t, err)

		// Verify edge still exists
		adjMap, err := dag.AdjacencyMap()
		assert.NoError(t, err)
		assert.Contains(t, adjMap[1], 2)
	})

	t.Run("returns error if adding edge fails for other reasons", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Try to add edge without adding vertices first
		// This should fail with an error other than ErrEdgeAlreadyExists
		err := dag.AddEdgeIfNotExist(1, 2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, graph.ErrVertexNotFound)
	})
}

// TestVertices tests the Vertices method
func TestVertices(t *testing.T) {
	t.Parallel()

	t.Run("returns empty slice for empty graph", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		vertices := dag.Vertices()
		assert.Empty(t, vertices)
	})

	t.Run("returns all vertices in the graph", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Add vertices
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))
		require.NoError(t, dag.AddVertex(3))

		// Get vertices
		vertices := dag.Vertices()

		// Verify all vertices are returned
		assert.ElementsMatch(t, []int{1, 2, 3}, vertices)
	})

	t.Run("returns vertices with complex types", func(t *testing.T) {
		t.Parallel()

		type complexType struct {
			ID   int
			Name string
		}

		dag := New(func(c complexType) int { return c.ID })

		// Add vertices with complex type
		v1 := complexType{ID: 1, Name: "one"}
		v2 := complexType{ID: 2, Name: "two"}

		require.NoError(t, dag.AddVertex(v1))
		require.NoError(t, dag.AddVertex(v2))

		// Get vertices
		vertices := dag.Vertices()

		// Verify all vertices are returned
		assert.ElementsMatch(t, []complexType{v1, v2}, vertices)
	})
}

// TestForEachVertex tests the ForEachVertex method
func TestForEachVertex(t *testing.T) {
	t.Parallel()

	t.Run("does nothing for empty graph", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		callCount := 0
		err := dag.ForEachVertex(func(_ int) error {
			callCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Zero(t, callCount)
	})

	t.Run("calls function for each vertex", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Add vertices
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))
		require.NoError(t, dag.AddVertex(3))

		// Track visited vertices
		visited := map[int]bool{}

		err := dag.ForEachVertex(func(i int) error {
			visited[i] = true
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, map[int]bool{1: true, 2: true, 3: true}, visited)
	})

	t.Run("returns error when function returns error", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Add vertices
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))
		require.NoError(t, dag.AddVertex(3))

		// Track visited vertices
		visited := map[int]bool{}
		expectedErr := errors.New("test error")

		err := dag.ForEachVertex(func(i int) error {
			visited[i] = true
			if i == 2 {
				return expectedErr
			}
			return nil
		})

		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, map[int]bool{1: true, 2: true}, visited)
	})
}

// TestInTopologicalOrder tests the InTopologicalOrder method
func TestInTopologicalOrder(t *testing.T) {
	t.Parallel()

	t.Run("does nothing for empty graph", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		callCount := 0
		err := dag.InTopologicalOrder(func(_ int) error {
			callCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Zero(t, callCount)
	})

	t.Run("calls function for each vertex in topological order", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))
		require.NoError(t, dag.AddVertex(3))
		require.NoError(t, dag.AddEdge(1, 2))
		require.NoError(t, dag.AddEdge(2, 3))

		// Track order of visited vertices
		var visited []int

		err := dag.InTopologicalOrder(func(i int) error {
			visited = append(visited, i)
			return nil
		})

		assert.NoError(t, err)
		assert.Len(t, visited, 3)

		// In topological order, we should visit 1, then 2, then 3
		assert.Equal(t, []int{1, 2, 3}, visited)
	})

	t.Run("stops iteration and returns error when function returns error", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))
		require.NoError(t, dag.AddVertex(3))
		require.NoError(t, dag.AddEdge(1, 2))
		require.NoError(t, dag.AddEdge(2, 3))

		// Track visited vertices
		var visited []int
		expectedErr := errors.New("test error")

		err := dag.InTopologicalOrder(func(i int) error {
			visited = append(visited, i)
			if i == 2 {
				return expectedErr
			}
			return nil
		})

		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, []int{1, 2}, visited)
	})
}

// TestInReverseTopologicalOrder tests the InReverseTopologicalOrder method
func TestInReverseTopologicalOrder(t *testing.T) {
	t.Parallel()

	t.Run("does nothing for empty graph", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		callCount := 0
		err := dag.InReverseTopologicalOrder(func(_ int) error {
			callCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Zero(t, callCount)
	})

	t.Run("calls function for each vertex in reverse topological order", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))
		require.NoError(t, dag.AddVertex(3))
		require.NoError(t, dag.AddEdge(1, 2))
		require.NoError(t, dag.AddEdge(2, 3))

		// Track order of visited vertices
		var visited []int

		err := dag.InReverseTopologicalOrder(func(i int) error {
			visited = append(visited, i)
			return nil
		})

		assert.NoError(t, err)
		assert.Len(t, visited, 3)

		// In reverse topological order, we should visit 3, then 2, then 1
		assert.Equal(t, []int{3, 2, 1}, visited)
	})

	t.Run("stops iteration and returns error when function returns error", func(t *testing.T) {
		t.Parallel()

		dag := New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, dag.AddVertex(1))
		require.NoError(t, dag.AddVertex(2))
		require.NoError(t, dag.AddVertex(3))
		require.NoError(t, dag.AddEdge(1, 2))
		require.NoError(t, dag.AddEdge(2, 3))

		// Track visited vertices
		var visited []int
		expectedErr := errors.New("test error")

		err := dag.InReverseTopologicalOrder(func(i int) error {
			visited = append(visited, i)
			if i == 2 {
				return expectedErr
			}
			return nil
		})

		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, []int{3, 2}, visited)
	})
}

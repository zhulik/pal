package dag_test

import (
	"errors"
	"testing"

	"github.com/zhulik/pal/pkg/dag"

	"github.com/dominikbraun/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hashFn(i int) int { return i }

// TestDAG_New tests the New function for DAG
func TestDAG_New(t *testing.T) {
	t.Parallel()

	t.Run("creates a new DAG with hash function", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		assert.NotNil(t, d)
		assert.NotNil(t, d.Graph)
	})
}

// TestDAG_AddVertexIfNotExist tests the AddVertexIfNotExist method of DAG
func TestDAG_AddVertexIfNotExist(t *testing.T) {
	t.Parallel()

	t.Run("adds a vertex if it doesn't exist", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		err := d.AddVertexIfNotExist(1)
		assert.NoError(t, err)

		// Verify vertex was added
		vertex, err := d.Vertex(1)
		assert.NoError(t, err)
		assert.Equal(t, 1, vertex)
	})

	t.Run("does nothing if vertex already exists", func(t *testing.T) {
		t.Parallel()

		dag := dag.New(hashFn)

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

// TestDAG_AddEdgeIfNotExist tests the AddEdgeIfNotExist method of DAG
func TestDAG_AddEdgeIfNotExist(t *testing.T) {
	t.Parallel()

	t.Run("adds an edge if it doesn't exist", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Add vertices first
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))

		// Add edge
		err := d.AddEdgeIfNotExist(1, 2)
		assert.NoError(t, err)

		// Verify edge was added
		adjMap, err := d.AdjacencyMap()
		assert.NoError(t, err)
		assert.Contains(t, adjMap[1], 2)
	})

	t.Run("does nothing if edge already exists", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Add vertices first
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))

		// Add edge first time
		err := d.AddEdge(1, 2)
		require.NoError(t, err)

		// Try to add same edge again
		err = d.AddEdgeIfNotExist(1, 2)
		assert.NoError(t, err)

		// Verify edge still exists
		adjMap, err := d.AdjacencyMap()
		assert.NoError(t, err)
		assert.Contains(t, adjMap[1], 2)
	})

	t.Run("returns error if adding edge fails for other reasons", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Try to add edge without adding vertices first
		// This should fail with an error other than ErrEdgeAlreadyExists
		err := d.AddEdgeIfNotExist(1, 2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, graph.ErrVertexNotFound)
	})
}

// TestDAG_Vertices tests the Vertices method of DAG
func TestDAG_Vertices(t *testing.T) {
	t.Parallel()

	t.Run("returns empty slice for empty graph", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		vertices := d.Vertices()
		assert.Empty(t, vertices)
	})

	t.Run("returns all vertices in the graph", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Add vertices
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))
		require.NoError(t, d.AddVertex(3))

		// Get vertices
		vertices := d.Vertices()

		// Verify all vertices are returned
		assert.ElementsMatch(t, []int{1, 2, 3}, vertices)
	})

	t.Run("returns vertices with complex types", func(t *testing.T) {
		t.Parallel()

		type complexType struct {
			ID   int
			Name string
		}

		d := dag.New(func(c complexType) int { return c.ID })

		// Add vertices with complex type
		v1 := complexType{ID: 1, Name: "one"}
		v2 := complexType{ID: 2, Name: "two"}

		require.NoError(t, d.AddVertex(v1))
		require.NoError(t, d.AddVertex(v2))

		// Get vertices
		vertices := d.Vertices()

		// Verify all vertices are returned
		assert.ElementsMatch(t, []complexType{v1, v2}, vertices)
	})
}

// TestDAG_ForEachVertex tests the ForEachVertex method of DAG
func TestDAG_ForEachVertex(t *testing.T) {
	t.Parallel()

	t.Run("does nothing for empty graph", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		callCount := 0
		err := d.ForEachVertex(func(_ int) error {
			callCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Zero(t, callCount)
	})

	t.Run("calls function for each vertex", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Add vertices
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))
		require.NoError(t, d.AddVertex(3))

		// Track visited vertices
		visited := map[int]bool{}

		err := d.ForEachVertex(func(i int) error {
			visited[i] = true
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, map[int]bool{1: true, 2: true, 3: true}, visited)
	})

	t.Run("returns error when function returns error", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Add vertices
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))
		require.NoError(t, d.AddVertex(3))

		// Track visited vertices
		visited := map[int]bool{}
		expectedErr := errors.New("test error")

		err := d.ForEachVertex(func(i int) error {
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

// TestDAG_InTopologicalOrder tests the InTopologicalOrder method of DAG
func TestDAG_InTopologicalOrder(t *testing.T) {
	t.Parallel()

	t.Run("does nothing for empty graph", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		callCount := 0
		err := d.InTopologicalOrder(func(_ int) error {
			callCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Zero(t, callCount)
	})

	t.Run("calls function for each vertex in topological order", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))
		require.NoError(t, d.AddVertex(3))
		require.NoError(t, d.AddEdge(1, 2))
		require.NoError(t, d.AddEdge(2, 3))

		// Track order of visited vertices
		var visited []int

		err := d.InTopologicalOrder(func(i int) error {
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

		d := dag.New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))
		require.NoError(t, d.AddVertex(3))
		require.NoError(t, d.AddEdge(1, 2))
		require.NoError(t, d.AddEdge(2, 3))

		// Track visited vertices
		var visited []int
		expectedErr := errors.New("test error")

		err := d.InTopologicalOrder(func(i int) error {
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

// TestDAG_InReverseTopologicalOrder tests the InReverseTopologicalOrder method of DAG
func TestDAG_InReverseTopologicalOrder(t *testing.T) {
	t.Parallel()

	t.Run("does nothing for empty graph", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		callCount := 0
		err := d.InReverseTopologicalOrder(func(_ int) error {
			callCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Zero(t, callCount)
	})

	t.Run("calls function for each vertex in reverse topological order", func(t *testing.T) {
		t.Parallel()

		d := dag.New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))
		require.NoError(t, d.AddVertex(3))
		require.NoError(t, d.AddEdge(1, 2))
		require.NoError(t, d.AddEdge(2, 3))

		// Track order of visited vertices
		var visited []int

		err := d.InReverseTopologicalOrder(func(i int) error {
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

		d := dag.New(hashFn)

		// Create a simple DAG: 1 -> 2 -> 3
		require.NoError(t, d.AddVertex(1))
		require.NoError(t, d.AddVertex(2))
		require.NoError(t, d.AddVertex(3))
		require.NoError(t, d.AddEdge(1, 2))
		require.NoError(t, d.AddEdge(2, 3))

		// Track visited vertices
		var visited []int
		expectedErr := errors.New("test error")

		err := d.InReverseTopologicalOrder(func(i int) error {
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

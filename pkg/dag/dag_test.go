package dag_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal/pkg/dag"
)

func TestNew(t *testing.T) {
	t.Parallel()

	d := dag.New[string, int]()
	assert.NotNil(t, d)
	assert.Equal(t, 0, d.VertexCount())
	assert.Equal(t, 0, d.EdgeCount())
}

func TestAddVertexIfNotExist(t *testing.T) {
	t.Parallel()

	t.Run("adds first vertex", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		assert.Equal(t, 1, d.VertexCount())
		assert.True(t, d.VertexExists("A"))

		val, exists := d.GetVertex("A")
		assert.True(t, exists)
		assert.Equal(t, 1, val)
	})

	t.Run("adds second vertex", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("B", 2)
		assert.Equal(t, 2, d.VertexCount())
		assert.True(t, d.VertexExists("A"))
		assert.True(t, d.VertexExists("B"))

		valA, existsA := d.GetVertex("A")
		assert.True(t, existsA)
		assert.Equal(t, 1, valA)

		valB, existsB := d.GetVertex("B")
		assert.True(t, existsB)
		assert.Equal(t, 2, valB)
	})

	t.Run("does not change existing vertex", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("A", 999)

		val, exists := d.GetVertex("A")
		assert.True(t, exists)
		assert.Equal(t, 1, val)
	})
}

func TestAddEdge(t *testing.T) {
	t.Parallel()

	t.Run("adds edge between existing vertices", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("B", 2)

		err := d.AddEdge("A", "B")
		assert.NoError(t, err)
		assert.True(t, d.EdgeExists("A", "B"))
		assert.Equal(t, 1, d.GetInDegree("B"))
		assert.Equal(t, 1, d.GetOutDegree("A"))
		assert.Equal(t, 1, d.EdgeCount())
	})

	t.Run("returns error for duplicate edge", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("B", 2)

		err := d.AddEdge("A", "B")
		assert.NoError(t, err)

		err = d.AddEdge("A", "B")
		assert.ErrorIs(t, err, dag.ErrEdgeAlreadyExists)
	})

	t.Run("does not allow adding edge from non-existent vertex", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		err := d.AddEdge("X", "Y")
		assert.ErrorIs(t, err, dag.ErrVertexNotFound)
	})

	t.Run("does not allow adding edge to non-existent vertex", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)

		err := d.AddEdge("A", "Y")
		assert.ErrorIs(t, err, dag.ErrVertexNotFound)
	})

	t.Run("prevents simple cycle", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("B", 2)
		d.AddVertexIfNotExist("C", 3)

		// Add edges that don't create cycles
		err := d.AddEdge("A", "B")
		assert.NoError(t, err)

		err = d.AddEdge("B", "C")
		assert.NoError(t, err)

		// Try to add edge that creates cycle
		err = d.AddEdge("C", "A")
		assert.ErrorIs(t, err, dag.ErrCycleDetected)

		// Verify the cycle-causing edge was not added
		assert.False(t, d.EdgeExists("C", "A"))
		assert.Equal(t, 0, d.GetInDegree("A"))
	})

	t.Run("prevents self-loop", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)

		err := d.AddEdge("A", "A")
		assert.ErrorIs(t, err, dag.ErrCycleDetected)
		assert.False(t, d.EdgeExists("A", "A"))
		assert.Equal(t, 0, d.GetInDegree("A"))
	})
}

func TestTopologicalOrder(t *testing.T) {
	t.Parallel()

	t.Run("simple chain", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("B", 2)
		d.AddVertexIfNotExist("C", 3)

		err := d.AddEdge("A", "B")
		assert.NoError(t, err)

		err = d.AddEdge("B", "C")
		assert.NoError(t, err)

		var result []string
		for id := range d.TopologicalOrder() {
			result = append(result, id)
		}

		assert.Equal(t, []string{"A", "B", "C"}, result)
	})

	t.Run("multiple paths", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		// Create a more complex DAG:
		//     A
		//    / \
		//   B   C
		//    \ /
		//     D
		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("B", 2)
		d.AddVertexIfNotExist("C", 3)
		d.AddVertexIfNotExist("D", 4)

		err := d.AddEdge("A", "B")
		assert.NoError(t, err)

		err = d.AddEdge("A", "C")
		assert.NoError(t, err)

		err = d.AddEdge("B", "D")
		assert.NoError(t, err)

		err = d.AddEdge("C", "D")
		assert.NoError(t, err)

		var result []string
		for id := range d.TopologicalOrder() {
			result = append(result, id)
		}

		Ai := slices.Index(result, "A")
		Bi := slices.Index(result, "B")
		Ci := slices.Index(result, "C")
		Di := slices.Index(result, "D")

		assert.True(t, Ai < Bi)
		assert.True(t, Ai < Ci)
		assert.True(t, Bi < Di)
		assert.True(t, Ci < Di)
	})

	t.Run("empty DAG", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		count := 0
		for range d.TopologicalOrder() {
			count++
		}

		assert.Equal(t, 0, count)
	})

	t.Run("single vertex", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)

		var result []string
		for id := range d.TopologicalOrder() {
			result = append(result, id)
		}

		assert.Len(t, result, 1)
		assert.Equal(t, "A", result[0])
	})
}

func TestReverseTopologicalOrder(t *testing.T) {
	t.Parallel()

	t.Run("simple chain", func(t *testing.T) {
		t.Parallel()
		d := dag.New[string, int]()

		d.AddVertexIfNotExist("A", 1)
		d.AddVertexIfNotExist("B", 2)
		d.AddVertexIfNotExist("C", 3)

		err := d.AddEdge("A", "B")
		assert.NoError(t, err)

		err = d.AddEdge("B", "C")
		assert.NoError(t, err)

		var result []string
		for id := range d.ReverseTopologicalOrder() {
			result = append(result, id)
		}

		assert.Equal(t, []string{"C", "B", "A"}, result)
	})
}

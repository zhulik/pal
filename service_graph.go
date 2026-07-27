package pal

import (
	"cmp"

	"github.com/zhulik/pal/internal/dag"
)

// ServiceGraph is the dependency graph type returned by [Container.Graph].
// It is a type alias of the internal DAG implementation so advanced users can
// still inspect and mutate the live graph without importing an internal package.
type ServiceGraph = dag.DAG[string, ServiceDef]

// Re-exported DAG errors for use with [errors.Is] / [errors.As] when mutating a [ServiceGraph].
var (
	ErrEdgeAlreadyExists = dag.ErrEdgeAlreadyExists
	ErrCycleDetected     = dag.ErrCycleDetected
	ErrVertexNotFound    = dag.ErrVertexNotFound
)

// CycleError is returned when adding an edge would introduce a cycle in a [ServiceGraph].
type CycleError[ID cmp.Ordered] = dag.CycleError[ID]

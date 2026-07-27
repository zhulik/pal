package inspect

import (
	"github.com/zhulik/pal"
)

// DAGJSON is an alias of [pal.TreeJSON] kept for compatibility.
type DAGJSON = pal.TreeJSON

// NodeJSON is an alias of [pal.TreeNodeJSON] kept for compatibility.
type NodeJSON = pal.TreeNodeJSON

// EdgeJSON is an alias of [pal.TreeEdgeJSON] kept for compatibility.
type EdgeJSON = pal.TreeEdgeJSON

// DAGToJSON encodes a dependency DAG as JSON.
// Advanced: prefer [pal.Pal.TreeJSON] when working from a running Pal instance.
func DAGToJSON(d *pal.ServiceGraph) ([]byte, error) {
	return pal.GraphToJSON(d)
}

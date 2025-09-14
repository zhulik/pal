package inspect

import (
	"encoding/json"
	"strings"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/pkg/dag"
)

type DAGJSON struct {
	Nodes []NodeJSON `json:"nodes"`
	Edges []EdgeJSON `json:"edges"`
}

type NodeJSON struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	InDegree  int    `json:"inDegree"`
	OutDegree int    `json:"outDegree"`

	Initer        bool `json:"initer"`
	Runner        bool `json:"runner"`
	HealthChecker bool `json:"healthChecker"`
	Shutdowner    bool `json:"shutdowner"`
}

type EdgeJSON struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func serviceToJSON(id string, inDegree int, outDegree int, service pal.ServiceDef) NodeJSON {
	var initer, runner, healthChecker, shutdowner bool

	if _, ok := service.Make().(pal.Initer); ok {
		initer = true
	}

	if _, ok := service.Make().(pal.Runner); ok {
		runner = true
	}

	if _, ok := service.Make().(pal.HealthChecker); ok {
		healthChecker = true
	}

	if _, ok := service.Make().(pal.Shutdowner); ok {
		shutdowner = true
	}

	idParts := strings.Split(id, "/")
	label := idParts[len(idParts)-1]

	if strings.HasPrefix(id, "*") {
		label = "*" + label
	}

	return NodeJSON{
		ID:        id,
		Label:     label,
		InDegree:  inDegree,
		OutDegree: outDegree,

		Initer:        initer,
		Runner:        runner,
		HealthChecker: healthChecker,
		Shutdowner:    shutdowner,
	}
}

func DAGToJSON(d *dag.DAG[string, pal.ServiceDef]) ([]byte, error) {
	var nodes []NodeJSON
	var edges []EdgeJSON

	// Convert all vertices to NodeJSON
	for id, service := range d.Vertices() {
		nodes = append(nodes, serviceToJSON(id, d.GetInDegree(id), len(d.Edges()[id]), service))
	}

	// Convert all edges to EdgeJSON
	for from, targets := range d.Edges() {
		for to := range targets {
			edges = append(edges, EdgeJSON{
				From: from,
				To:   to,
			})
		}
	}

	dagJSON := DAGJSON{
		Nodes: nodes,
		Edges: edges,
	}

	return json.Marshal(dagJSON)
}

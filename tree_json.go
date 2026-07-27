package pal

import (
	"encoding/json"
	"strings"
)

// TreeJSON is the JSON shape of a dependency graph (nodes and edges).
// Advanced: useful for custom visualizers; [inspect] uses [Pal.TreeJSON].
type TreeJSON struct {
	Nodes []TreeNodeJSON `json:"nodes"`
	Edges []TreeEdgeJSON `json:"edges"`
}

// TreeNodeJSON describes one service vertex in [TreeJSON].
type TreeNodeJSON struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	InDegree  int    `json:"inDegree"`
	OutDegree int    `json:"outDegree"`

	Initer        bool `json:"initer"`
	Runner        bool `json:"runner"`
	RunConfiger   bool `json:"runConfiger"`
	HealthChecker bool `json:"healthChecker"`
	Shutdowner    bool `json:"shutdowner"`
}

// TreeEdgeJSON describes one dependency edge in [TreeJSON].
type TreeEdgeJSON struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func serviceToTreeNodeJSON(id string, inDegree int, outDegree int, service ServiceDef) TreeNodeJSON {
	var initer, runner, runConfiger, healthChecker, shutdowner bool

	instance := service.Make()

	if _, ok := instance.(PalIniter); ok {
		initer = true
	} else if _, ok := instance.(Initer); ok {
		initer = true
	}

	if _, ok := instance.(PalRunner); ok {
		runner = true
	} else if _, ok := instance.(Runner); ok {
		runner = true
	}

	if _, ok := instance.(PalRunConfiger); ok {
		runConfiger = true
	} else if _, ok := instance.(RunConfiger); ok {
		runConfiger = true
	}

	if _, ok := instance.(PalHealthChecker); ok {
		healthChecker = true
	} else if _, ok := instance.(HealthChecker); ok {
		healthChecker = true
	}

	if _, ok := instance.(PalShutdowner); ok {
		shutdowner = true
	} else if _, ok := instance.(Shutdowner); ok {
		shutdowner = true
	}

	idParts := strings.Split(id, "/")
	label := idParts[len(idParts)-1]

	if strings.HasPrefix(id, "*") {
		label = "*" + label
	}

	return TreeNodeJSON{
		ID:        id,
		Label:     label,
		InDegree:  inDegree,
		OutDegree: outDegree,

		Initer:        initer,
		Runner:        runner,
		RunConfiger:   runConfiger,
		HealthChecker: healthChecker,
		Shutdowner:    shutdowner,
	}
}

// GraphToJSON encodes a dependency DAG as JSON.
// Advanced: prefer [Pal.TreeJSON] unless you already hold a [ServiceGraph].
func GraphToJSON(d *ServiceGraph) ([]byte, error) {
	var nodes []TreeNodeJSON
	var edges []TreeEdgeJSON

	for id, service := range d.Vertices() {
		nodes = append(nodes, serviceToTreeNodeJSON(id, d.GetInDegree(id), len(d.Edges()[id]), service))
	}

	for from, targets := range d.Edges() {
		for to := range targets {
			edges = append(edges, TreeEdgeJSON{
				From: from,
				To:   to,
			})
		}
	}

	return json.Marshal(TreeJSON{
		Nodes: nodes,
		Edges: edges,
	})
}

// TreeJSON returns a JSON encoding of this Pal instance's dependency graph.
func (p *Pal) TreeJSON() ([]byte, error) {
	return GraphToJSON(p.container.Graph())
}

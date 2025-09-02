package draw

import (
	"fmt"
	"strings"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/pkg/dag"
)

type dorRenderer struct {
	*strings.Builder

	graph *dag.DAG[string, pal.ServiceDef]
}

func (r *dorRenderer) Render() []byte {
	r.WriteString(`
	digraph DependencyGraph {
		graph [size="15,30"];
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif"]
		edge [fontname="Helvetica,Arial,sans-serif"]
	`)

	for _, vertex := range r.graph.TopologicalOrder() {
		r.renderVertex(vertex)
	}

	for id, vertex := range r.graph.TopologicalOrder() {
		for _, edge := range r.graph.OutEdges(id) {
			edgeVertex, _ := r.graph.GetVertex(edge)
			r.renderEdge(vertex, edgeVertex)
		}
	}

	r.WriteString("}")

	return []byte(r.String())
}

func (r *dorRenderer) vertexShape(vertex pal.ServiceDef) string {
	if _, ok := vertex.Make().(pal.Runner); ok {
		return "cds"
	}

	if len(r.graph.OutEdges(vertex.Name())) == 0 {
		return "house"
	}

	return "ellipse"
}

func (r *dorRenderer) edgeStyle(source, target pal.ServiceDef) string {
	if strings.HasPrefix(target.Name(), "*") && strings.HasPrefix(source.Name(), "*") {
		return "solid"
	}

	if strings.HasPrefix(target.Name(), "*") || strings.HasPrefix(source.Name(), "*") {
		return "dashed"
	}
	return "dotted"
}

func (r *dorRenderer) vertexColor(vertex pal.ServiceDef) string {
	_, runner := vertex.Make().(pal.Runner)
	if !runner && len(r.graph.InEdges(vertex.Name())) == 0 {
		return "indianred1"
	}
	return "transparent"
}

func (r *dorRenderer) renderVertex(vertex pal.ServiceDef) {
	fmt.Fprintf(r, `%s [label="%s", tooltip="%s", shape="%s", fillcolor="%s", style="filled"]`,
		sanitizeID(vertex.Name()), simplifyName(vertex.Name()), vertex.Name(), r.vertexShape(vertex), r.vertexColor(vertex),
	)
	r.WriteRune('\n')
}

func (r *dorRenderer) renderEdge(source, target pal.ServiceDef) {
	fmt.Fprintf(r, `%s -> %s [style="%s"]`, sanitizeID(source.Name()), sanitizeID(target.Name()), r.edgeStyle(source, target))
	r.WriteRune('\n')
}

func RenderDOT(graph *dag.DAG[string, pal.ServiceDef]) []byte {
	renderer := &dorRenderer{Builder: &strings.Builder{}, graph: graph}

	return renderer.Render()
}

func sanitizeID(id string) string {
	result := strings.NewReplacer("*", "", ".", "_", "/", "_", "-", "_").Replace(id)
	return result
}

func simplifyName(name string) string {
	parts := strings.Split(name, "/")
	last := parts[len(parts)-1]
	if strings.HasPrefix(name, "*") {
		return "*" + last
	}
	return last
}

package draw

import (
	"fmt"
	"strings"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/pkg/dag"
)

func sanitizeID(id string) string {
	withoutStars := strings.ReplaceAll(id, "*", "")
	return strings.ReplaceAll(withoutStars, ".", "_")
}

func shape(vertex pal.ServiceDef) string {
	if _, ok := vertex.Make().(pal.Runner); ok {
		return "cds"
	}
	// if vertex.Arguments() == 0 {
	// 	return "box"
	// }
	return "ellipse"
}

func renderVertex(builder *strings.Builder, vertex pal.ServiceDef) {
	builder.WriteString(fmt.Sprintf(`%s [label="%s", shape="%s"]`, sanitizeID(vertex.Name()), vertex.Name(), shape(vertex)))
	builder.WriteRune('\n')
}

func RenderDOT[ID comparable, T pal.ServiceDef](graph *dag.DAG[ID, T]) string {
	builder := &strings.Builder{}

	builder.WriteString(`
	digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif"]
		edge [fontname="Helvetica,Arial,sans-serif"]
	`)

	for _, vertex := range graph.TopologicalOrder() {
		renderVertex(builder, vertex)
	}

	for id, vertex := range graph.TopologicalOrder() {
		for _, edge := range graph.OutEdges(id) {
			edgeVertex, _ := graph.GetVertex(edge)
			builder.WriteString(fmt.Sprintf(`%s -> %s`, sanitizeID(vertex.Name()), sanitizeID(edgeVertex.Name())))
			builder.WriteRune('\n')
		}
	}

	builder.WriteString("}")

	return builder.String()
}

package inspect

import (
	"context"

	"github.com/goccy/go-graphviz"
)

type Graphviz struct {
	*graphviz.Graphviz
}

func (g *Graphviz) Shutdown(_ context.Context) error {
	return g.Close()
}

func (g *Graphviz) Init(ctx context.Context) error {
	var err error

	g.Graphviz, err = graphviz.New(ctx)
	return err
}

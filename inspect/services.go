package inspect

import (
	"context"

	"github.com/goccy/go-graphviz"
	"github.com/zhulik/pal"
)

func ProvideBase() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide(&Inspect{}),
		pal.ProvideFactory0(func(context.Context) (*VM, error) {
			return &VM{}, nil
		}),

		pal.ProvideFn(graphviz.New).
			ToShutdown(func(_ context.Context, g *graphviz.Graphviz, _ *pal.Pal) error {
				return g.Close()
			}),
	)
}

func ProvideRemoteConsole() pal.ServiceDef {
	return pal.ProvideRunner(RemoteConsole)
}

func ProvideConsole() pal.ServiceDef {
	return pal.Provide(&Console{})
}

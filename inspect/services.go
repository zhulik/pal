package inspect

import (
	"context"

	"github.com/zhulik/pal"
)

func ProvideBase() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide(&Inspect{}),
		pal.ProvideFactory0(func(context.Context) (*VM, error) {
			return &VM{}, nil
		}),
	)
}

func ProvideRemoteConsole() pal.ServiceDef {
	return pal.ProvideRunner(RemoteConsole)
}

func ProvideConsole() pal.ServiceDef {
	return pal.Provide(&Console{})
}

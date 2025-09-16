package inspect

import (
	"github.com/zhulik/pal"
)

const (
	defaultPort = 24242
)

func Provide(port ...int) pal.ServiceDef {
	p := defaultPort
	if len(port) > 0 {
		p = port[0]
	}

	return pal.ProvideList(
		pal.Provide(&Inspect{port: p}),
	)
}

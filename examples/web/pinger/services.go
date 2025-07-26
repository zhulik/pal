package pinger

import (
	"github.com/zhulik/pal"
	"github.com/zhulik/pal/examples/web/core"
)

// Provide provides the pinger service.
func Provide() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide[core.Pinger](&Pinger{}),
		// Add more services here, if needed.
	)
}

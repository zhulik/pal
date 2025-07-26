package server

import (
	"github.com/zhulik/pal"
)

// Provide provides the server service.
func Provide() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide(&Server{}),
		// Add more services here, if needed.
	)
}

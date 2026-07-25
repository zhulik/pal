package app

import (
	"context"
	"time"

	"github.com/zhulik/pal"
)

func Run(ctx context.Context, services ...pal.ServiceDef) error {
	p := pal.New(
		services...,
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(3 * time.Second)
	return p.Run(ctx)
}

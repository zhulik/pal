package main

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/zhulik/pal"
)

type Heartbeater interface {
	Heartbeat(ctx context.Context) error
}

type heartbeater struct {
	count int
}

func (h *heartbeater) Heartbeat(ctx context.Context) error {
	slog.InfoContext(ctx, "Heartbeat", "count", h.count)
	return nil
}

func mainRunner(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting runner")
	c := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			c++
			pal.MustInvokeByInterface[Heartbeater](ctx, nil, c).Heartbeat(ctx)
		}
	}
}

func TestFactoryExperiment(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := pal.New(
		pal.ProvideRunner(mainRunner),
		pal.ProvideFactoryExperement1[Heartbeater](func(ctx context.Context, count int) (*heartbeater, error) {
			return &heartbeater{count: count}, nil
		}),
	).
		InjectSlog().
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second).
		Run(t.Context())
	if err != nil {
		panic(err)
	}
}

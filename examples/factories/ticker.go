package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/zhulik/pal"
)

// ticker is a concrete implementation of the ticker interface.
type ticker struct {
	Logger *slog.Logger // logger is injected by pal as is

	Pal *pal.Pal

	// CreatePinger is a factory function that creates a pinger service, it is injected by pal.
	CreatePinger func(ctx context.Context, url string) (Pinger, error)

	pinger Pinger
	ticker *time.Ticker // ticker is created in Init and stopped in Shutdown.
}

// Init initializes the ticker service.
func (t *ticker) Init(ctx context.Context) error { //nolint:unparam
	defer t.Logger.Info("ticker initialized")

	pinger, err := t.CreatePinger(ctx, "https://google.com")
	if err != nil {
		return err
	}
	t.pinger = pinger

	t.ticker = time.NewTicker(time.Second)

	return nil
}

// Shutdown closes the ticker service.
func (t *ticker) Shutdown(_ context.Context) error { //nolint:unparam
	t.Logger.Info("ticker shut down")

	t.ticker.Stop()

	return nil
}

// Run runs the ticker service, calls Pinger.Ping every second.
func (t *ticker) Run(ctx context.Context) error { //nolint:unparam
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-t.ticker.C:
			if err := t.pinger.Ping(ctx); err != nil {
				t.Logger.Error("Failed to ping", "error", err)
			}
		}
	}
}

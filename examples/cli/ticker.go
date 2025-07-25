package main

import (
	"context"
	"log/slog"
	"time"
)

// ticker is a concrete implementation of the ticker interface.
type ticker struct {
	Pinger Pinger       // pinger is injected by pal, using the Pinger interface.
	Logger *slog.Logger // logger is injected by pal as is

	ticker *time.Ticker // ticker is created in Init and stopped in Shutdown.
}

// Init initializes the ticker service.
func (t *ticker) Init(_ context.Context) error { //nolint:unparam
	t.Logger.Info("ticker initialized")

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
			if err := t.Pinger.Ping(ctx); err != nil {
				t.Logger.Error("Failed to ping", "error", err)
			}
		}
	}
}

package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/zhulik/pal"
)

type Pinger struct {
	client *http.Client
}

func (p *Pinger) Init(_ context.Context) error {
	defer slog.Info("Pinger initialized")

	p.client = &http.Client{
		Timeout: 5 * time.Second,
	}

	return nil
}

func (p *Pinger) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET", "https://google.com", nil)
			if err != nil {
				slog.Error("Failed to create request", "error", err)
				return err
			}
			resp, err := p.client.Do(req)
			if err != nil {
				return nil
			}
			slog.Info("GET google.com", "status", resp.Status)
			resp.Body.Close()
		}
	}
}

func (p *Pinger) Shutdown(_ context.Context) error {
	time.Sleep(2 * time.Second)

	defer slog.Info("Pinger shut down")
	p.client.CloseIdleConnections()

	return nil
}

func main() {
	p := pal.New(
		pal.Provide(&Pinger{}),
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(3 * time.Second)

	if err := p.Run(context.Background()); err != nil {
		slog.Error("Error running pal", "error", err)
	}
}

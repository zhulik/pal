package pinger

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Pinger is a concrete implementation of the Pinger interface.
type Pinger struct {
	Logger *slog.Logger

	client *http.Client
}

// Init initializes the pinger service, creates a http client.
func (p *Pinger) Init(_ context.Context) error {
	defer p.Logger.Info("Pinger initialized")

	p.client = &http.Client{
		Timeout: 5 * time.Second,
	}

	return nil
}

// Shutdown closes the http client.
func (p *Pinger) Shutdown(_ context.Context) error {
	defer p.Logger.Info("Pinger shut down")
	p.client.CloseIdleConnections()

	return nil
}

// Ping pings google.com.
func (p *Pinger) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://google.com", nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	p.Logger.Info("GET google.com", "status", resp.Status)
	return resp.Body.Close()
}

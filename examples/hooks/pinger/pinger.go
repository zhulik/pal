package pinger

import (
	"context"
	"log/slog"
	"net/http"
)

// Pinger is a concrete implementation of the Pinger interface.
type Pinger struct {
	Logger *slog.Logger

	client *http.Client
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

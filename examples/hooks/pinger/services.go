package pinger

import (
	"context"
	"net/http"
	"time"

	"github.com/zhulik/pal"
)

// Provide provides the pinger service.
func Provide() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide(&Pinger{}).
			ToInit(func(ctx context.Context, pinger *Pinger, pal *pal.Pal) error {
				defer pinger.Logger.Info("Pinger initialized")

				pinger.client = &http.Client{
					Timeout: 5 * time.Second,
				}

				return nil
			}).
			ToShutdown(func(ctx context.Context, pinger *Pinger, pal *pal.Pal) error {
				defer pinger.Logger.Info("Pinger shut down")
				pinger.client.CloseIdleConnections()

				return nil
			}),
		// Add more services here, if needed.
	)
}

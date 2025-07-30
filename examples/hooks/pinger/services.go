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
			ToInit(func(_ context.Context, pinger *Pinger, _ *pal.Pal) error {
				defer pinger.Logger.Info("Pinger initialized")

				pinger.client = &http.Client{
					Timeout: 5 * time.Second,
				}

				return nil
			}).
			ToShutdown(func(_ context.Context, pinger *Pinger, _ *pal.Pal) error {
				defer pinger.Logger.Info("Pinger shut down")
				pinger.client.CloseIdleConnections()

				return nil
			}),
		// Add more services here, if needed.
	)
}

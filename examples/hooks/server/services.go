package server

import (
	"context"
	"net/http"

	"github.com/zhulik/pal"
)

// Provide provides the server service.
func Provide() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide(&Server{}).ToInit(func(_ context.Context, server *Server, _ *pal.Pal) error {
			defer server.Logger.Info("Server initialized")

			server.server = &http.Server{ //nolint:gosec
				Addr: ":8080",
			}

			server.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := server.Pinger.Ping(r.Context()); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("pong")) //nolint:errcheck
			})

			return nil
		}),
		// Add more services here, if needed.
	)
}

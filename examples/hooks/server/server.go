package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/zhulik/pal/examples/web/core"
)

// Server is a simple http Server that calls Pinger.Ping on each request.
type Server struct {
	Pinger core.Pinger
	Logger *slog.Logger

	server *http.Server
}

// Run runs the server.
func (s *Server) Run(ctx context.Context) error {
	s.Logger.Info("Server running on :8080. Do `curl http://localhost:8080/` to see it in action.")

	// We don't use Shutdown here because ListenAndServe() does not natively support context.
	// instead we use a goroutine to listen for the context done signal and shutdown the server.
	go func() {
		<-ctx.Done()
		s.server.Shutdown(context.Background()) //nolint:errcheck
	}()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

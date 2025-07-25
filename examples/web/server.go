package main

import (
	"context"
	"log/slog"
	"net/http"
)

// server is a simple http server that calls Pinger.Ping on each request.
type server struct {
	Pinger Pinger
	Logger *slog.Logger

	server *http.Server
}

// Init initializes the server.
func (s *server) Init(_ context.Context) error {
	defer s.Logger.Info("Server initialized")

	s.server = &http.Server{ //nolint:gosec
		Addr: ":8080",
	}

	s.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.Pinger.Ping(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong")) //nolint:errcheck
	})

	return nil
}

// Run runs the server.
func (s *server) Run(ctx context.Context) error {
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

package pal

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type palHealthCheckServer interface {
	handle(w http.ResponseWriter, r *http.Request)
}

type healthCheckServer struct {
	Pal    *Pal
	Logger *slog.Logger

	server *http.Server
}

func (h *healthCheckServer) RunConfig() *RunConfig {
	return &RunConfig{
		Wait: false,
	}
}

func (h *healthCheckServer) Init(_ context.Context) error {
	h.server = &http.Server{
		Addr:              h.Pal.config.HealthCheckAddr,
		Handler:           http.HandlerFunc(h.handle),
		ReadHeaderTimeout: time.Second,
	}

	return nil
}

func (h healthCheckServer) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != h.Pal.config.HealthCheckPath {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.Pal.HealthCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *healthCheckServer) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		h.server.Shutdown(context.Background()) //nolint:errcheck
	}()

	h.Logger.Info("Health check server running", "addr", h.server.Addr)
	err := h.server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

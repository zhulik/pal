package pal

import (
	"context"
	"net/http"
	"time"
)

type palHealthCheckServer interface {
	handle(w http.ResponseWriter, r *http.Request)
}

type healthCheckServer struct {
	Pal *Pal

	addr string
	path string

	server *http.Server
}

func (h *healthCheckServer) RunConfig() *RunConfig {
	return &RunConfig{
		Wait: false,
	}
}

func (h *healthCheckServer) Init(_ context.Context) error {
	h.server = &http.Server{
		Addr:              h.addr,
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

	if r.URL.Path != h.path {
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

		// create a new context as the one passed to Run is already canceled
		ctx, cancel := context.WithTimeout(context.Background(), h.Pal.Config().ShutdownTimeout)
		defer cancel()
		h.server.Shutdown(ctx) //nolint:errcheck
	}()

	err := h.server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

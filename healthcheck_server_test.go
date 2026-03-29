package pal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckServer_RunConfig(t *testing.T) {
	t.Parallel()

	h := &healthCheckServer{}
	cfg := h.RunConfig()

	require.NotNil(t, cfg)
	assert.False(t, cfg.Wait)
}

func TestHealthCheckServer_Init(t *testing.T) {
	t.Parallel()

	p := New()
	h := &healthCheckServer{
		Pal:  p,
		addr: "127.0.0.1:0",
		path: "/healthz",
	}

	require.NoError(t, h.Init(t.Context()))
	require.NotNil(t, h.server)
	assert.Equal(t, h.addr, h.server.Addr)
	assert.Equal(t, time.Second, h.server.ReadHeaderTimeout)
	assert.NotNil(t, h.server.Handler)
}

func TestHealthCheckServer_Run_shutdownOnContextCancel(t *testing.T) {
	t.Parallel()

	p := New().ShutdownTimeout(5 * time.Second)
	h := &healthCheckServer{
		Pal:  p,
		addr: "127.0.0.1:0",
		path: "/healthz",
	}

	require.NoError(t, h.Init(t.Context()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}

type healthCheckOKA struct{}

func (*healthCheckOKA) HealthCheck(context.Context) error { return nil }

type healthCheckOKB struct{}

func (*healthCheckOKB) HealthCheck(context.Context) error { return nil }

type healthCheckFail struct{}

func (*healthCheckFail) HealthCheck(context.Context) error { return errors.New("unhealthy") }

func TestHealthCheckServer_handle_allServicesHealthy(t *testing.T) {
	t.Parallel()

	p := New(
		Provide(&healthCheckOKA{}),
		Provide(&healthCheckOKB{}),
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)

	require.NoError(t, p.Init(t.Context()))

	h := &healthCheckServer{Pal: p, path: "/healthz"}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.handle(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthCheckServer_handle_someServicesUnhealthy(t *testing.T) {
	t.Parallel()

	p := New(
		Provide(&healthCheckOKA{}),
		Provide(&healthCheckFail{}),
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)

	require.NoError(t, p.Init(t.Context()))

	h := &healthCheckServer{Pal: p, path: "/healthz"}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.handle(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHealthCheckServer_handle_wrongMethod(t *testing.T) {
	t.Parallel()

	h := &healthCheckServer{Pal: New(), path: "/healthz"}
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.handle(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHealthCheckServer_handle_wrongPath(t *testing.T) {
	t.Parallel()

	h := &healthCheckServer{Pal: New(), path: "/healthz"}
	req := httptest.NewRequest(http.MethodGet, "/wrong", nil)
	rec := httptest.NewRecorder()

	h.handle(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

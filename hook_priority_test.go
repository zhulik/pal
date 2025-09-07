package pal_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
)

// HookPriorityTestService implements all lifecycle interfaces for testing
type HookPriorityTestService struct {
	initCalled        bool
	healthCheckCalled bool
	shutdownCalled    bool
	runCalled         bool
	initError         error
	healthCheckError  error
	shutdownError     error
	runError          error
}

func (h *HookPriorityTestService) Init(_ context.Context) error {
	h.initCalled = true
	return h.initError
}

func (h *HookPriorityTestService) HealthCheck(_ context.Context) error {
	h.healthCheckCalled = true
	return h.healthCheckError
}

func (h *HookPriorityTestService) Shutdown(_ context.Context) error {
	h.shutdownCalled = true
	return h.shutdownError
}

func (h *HookPriorityTestService) Run(_ context.Context) error {
	h.runCalled = true
	return h.runError
}

// TestHookPriority_ToInit tests that ToInit hook has priority over Init method
func TestHookPriority_ToInit(t *testing.T) {
	t.Parallel()

	t.Run("ToInit hook is called instead of Init method", func(t *testing.T) {
		t.Parallel()

		service := &HookPriorityTestService{}

		hookCalled := false
		palService := pal.Provide(service).
			ToInit(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				hookCalled = true
				return nil
			})

		p := newPal(palService)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Verify hook was called
		assert.True(t, hookCalled, "ToInit hook should have been called")

		// Verify Init method was NOT called
		assert.False(t, service.initCalled, "Init method should not be called when ToInit hook is specified")
	})

	t.Run("ToInit hook error is propagated", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("init hook error")
		service := &HookPriorityTestService{}

		palService := pal.Provide(service).
			ToInit(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				return expectedErr
			})

		p := newPal(palService)

		err := p.Init(t.Context())
		assert.ErrorIs(t, err, expectedErr, "ToInit hook error should be propagated")
	})

	t.Run("Init method is called when no ToInit hook is specified", func(t *testing.T) {
		t.Parallel()

		service := &HookPriorityTestService{}

		palService := pal.Provide(service)
		p := newPal(palService)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Verify Init method was called
		assert.True(t, service.initCalled, "Init method should be called when no ToInit hook is specified")
	})
}

// TestHookPriority_ToHealthCheck tests that ToHealthCheck hook has priority over HealthCheck method
func TestHookPriority_ToHealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("ToHealthCheck hook is called instead of HealthCheck method", func(t *testing.T) {
		t.Parallel()

		service := &HookPriorityTestService{}

		hookCalled := false
		palService := pal.Provide(service).
			ToHealthCheck(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				hookCalled = true
				return nil
			})

		p := newPal(palService)
		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		// Initialize first
		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Perform health check
		err = p.HealthCheck(ctx)
		assert.NoError(t, err)

		// Verify hook was called
		assert.True(t, hookCalled, "ToHealthCheck hook should have been called")

		// Verify HealthCheck method was NOT called
		assert.False(t, service.healthCheckCalled, "HealthCheck method should not be called when ToHealthCheck hook is specified")
	})

	t.Run("ToHealthCheck hook error is propagated", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("health check hook error")
		service := &HookPriorityTestService{}

		palService := pal.Provide(service).
			ToHealthCheck(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				return expectedErr
			})

		p := newPal(palService)

		// Initialize first
		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Perform health check
		err = p.HealthCheck(t.Context())
		assert.ErrorIs(t, err, expectedErr, "ToHealthCheck hook error should be propagated")
	})

	t.Run("HealthCheck method is called when no ToHealthCheck hook is specified", func(t *testing.T) {
		t.Parallel()

		service := &HookPriorityTestService{}

		palService := pal.Provide(service)
		p := newPal(palService)

		// Initialize first
		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Perform health check
		err = p.HealthCheck(t.Context())
		assert.NoError(t, err)

		// Verify HealthCheck method was called
		assert.True(t, service.healthCheckCalled, "HealthCheck method should be called when no ToHealthCheck hook is specified")
	})
}

// TestHookPriority_ToShutdown tests that ToShutdown hook has priority over Shutdown method
func TestHookPriority_ToShutdown(t *testing.T) {
	t.Parallel()

	t.Run("ToShutdown hook is called instead of Shutdown method", func(t *testing.T) {
		t.Skip("TODO: this test is not working as expected")
		t.Parallel()

		service := &HookPriorityTestService{}

		hookCalled := false
		palService := pal.Provide(service).
			ToShutdown(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				hookCalled = true
				return nil
			})

		p := newPal(palService)

		// Initialize first
		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Shutdown
		p.Shutdown()
		err = p.Run(t.Context())
		assert.NoError(t, err)

		// Verify hook was called
		assert.True(t, hookCalled, "ToShutdown hook should have been called")

		// Verify Shutdown method was NOT called
		assert.False(t, service.shutdownCalled, "Shutdown method should not be called when ToShutdown hook is specified")
	})

	t.Run("ToShutdown hook error is propagated", func(t *testing.T) {
		t.Skip("TODO: this test is not working as expected")
		t.Parallel()

		expectedErr := errors.New("shutdown hook error")
		service := &HookPriorityTestService{}

		palService := pal.Provide(service).
			ToShutdown(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				return expectedErr
			})

		p := newPal(palService)

		// Initialize first
		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Shutdown
		p.Shutdown()
		err = p.Run(t.Context())
		assert.ErrorIs(t, err, expectedErr, "ToShutdown hook error should be propagated")
	})

	t.Run("Shutdown method is called when no ToShutdown hook is specified", func(t *testing.T) {
		t.Skip("TODO: this test is not working as expected")
		t.Parallel()

		service := &HookPriorityTestService{}

		palService := pal.Provide(service)
		p := newPal(palService)

		// Initialize first
		err := p.Init(t.Context())
		assert.NoError(t, err)

		// Shutdown
		p.Shutdown()
		err = p.Run(t.Context())
		assert.NoError(t, err)

		// Verify Shutdown method was called
		assert.True(t, service.shutdownCalled, "Shutdown method should be called when no ToShutdown hook is specified")
	})
}

// TestHookPriority_MultipleHooks tests that multiple hooks can be used together
func TestHookPriority_MultipleHooks(t *testing.T) {
	t.Parallel()

	t.Run("all hooks are called and methods are not", func(t *testing.T) {
		t.Skip("TODO: this test is not working as expected")
		t.Parallel()

		service := &HookPriorityTestService{}

		initHookCalled := false
		healthCheckHookCalled := false
		shutdownHookCalled := false

		palService := pal.Provide(service).
			ToInit(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				initHookCalled = true
				return nil
			}).
			ToHealthCheck(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				healthCheckHookCalled = true
				return nil
			}).
			ToShutdown(func(_ context.Context, _ *HookPriorityTestService, _ *pal.Pal) error {
				shutdownHookCalled = true
				return nil
			})

		p := newPal(palService)
		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		// Initialize
		err := p.Init(t.Context())
		assert.NoError(t, err)
		assert.True(t, initHookCalled, "ToInit hook should have been called")

		// Health check
		err = p.HealthCheck(ctx)
		assert.NoError(t, err)
		assert.True(t, healthCheckHookCalled, "ToHealthCheck hook should have been called")

		// Shutdown
		p.Shutdown()
		err = p.Run(t.Context())
		assert.NoError(t, err)
		assert.True(t, shutdownHookCalled, "ToShutdown hook should have been called")

		// Verify none of the interface methods were called
		assert.False(t, service.initCalled, "Init method should not be called when ToInit hook is specified")
		assert.False(t, service.healthCheckCalled, "HealthCheck method should not be called when ToHealthCheck hook is specified")
		assert.False(t, service.shutdownCalled, "Shutdown method should not be called when ToShutdown hook is specified")
	})
}

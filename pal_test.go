package pal_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zhulik/pal"
	"github.com/zhulik/pal/pkg/core"
)

// MockLogger is a mock implementation of the core.LoggerFn
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(fmt string, args ...any) {
	m.Called(fmt, args)
}

// TestPal_New tests the New function
func TestPal_New(t *testing.T) {
	t.Parallel()

	t.Run("creates a new Pal instance with no services", func(t *testing.T) {
		t.Parallel()

		p := pal.New()

		assert.NotNil(t, p)
		assert.Empty(t, p.Services())
	})

	t.Run("creates a new Pal instance with services", func(t *testing.T) {
		t.Parallel()

		p := pal.New(
			pal.Provide[TestInterface, TestStruct](),
		)

		assert.NoError(t, p.Init(t.Context()))
	})
}

// TestPal_FromContext tests the FromContext function
func TestPal_FromContext(t *testing.T) {
	t.Parallel()

	t.Run("retrieves Pal from context", func(t *testing.T) {
		t.Parallel()

		p := pal.New()
		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		result := pal.FromContext(ctx)

		assert.Same(t, p, result)
	})
}

// TestPal_InitTimeout tests the InitTimeout method
func TestPal_InitTimeout(t *testing.T) {
	t.Parallel()

	t.Run("sets the init timeout", func(t *testing.T) {
		t.Parallel()

		p := pal.New()
		timeout := 5 * time.Second

		result := p.InitTimeout(timeout)

		assert.Same(t, p, result) // Method should return the Pal instance for chaining
	})
}

// TestPal_HealthCheckTimeout tests the HealthCheckTimeout method
func TestPal_HealthCheckTimeout(t *testing.T) {
	t.Parallel()

	t.Run("sets the health check timeout", func(t *testing.T) {
		t.Parallel()

		p := pal.New()
		timeout := 5 * time.Second

		result := p.HealthCheckTimeout(timeout)

		assert.Same(t, p, result) // Method should return the Pal instance for chaining
	})
}

// TestPal_ShutdownTimeout tests the ShutdownTimeout method
func TestPal_ShutdownTimeout(t *testing.T) {
	t.Parallel()

	t.Run("sets the shutdown timeout", func(t *testing.T) {
		t.Parallel()

		p := pal.New()
		timeout := 5 * time.Second

		result := p.ShutdownTimeout(timeout)

		assert.Same(t, p, result) // Method should return the Pal instance for chaining
	})
}

// TestPal_SetLogger tests the SetLogger method
func TestPal_SetLogger(t *testing.T) {
	t.Parallel()

	t.Run("sets the logger", func(t *testing.T) {
		t.Parallel()

		p := pal.New()
		logger := func(string, ...any) {}

		result := p.SetLogger(logger)

		assert.Same(t, p, result) // Method should return the Pal instance for chaining
	})
}

// TestPal_HealthCheck tests the HealthCheck method
func TestPal_HealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("performs health check on all services", func(t *testing.T) {
		t.Parallel()

		// Create a service that implements HealthChecker
		service := pal.Provide[TestInterface, TestStruct]()
		p := pal.New(service)

		err := p.HealthCheck(t.Context())

		assert.NoError(t, err)
	})

	// TODO: health check times out
}

// TestPal_Shutdown tests the Shutdown method
func TestPal_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("schedules shutdown with no errors", func(t *testing.T) {
		t.Parallel()

		p := pal.New()

		// This is a non-blocking call
		p.Shutdown()

		// No way to directly test the effect, but we can verify it doesn't panic
	})

	t.Run("schedules shutdown with errors", func(t *testing.T) {
		t.Parallel()

		p := pal.New()
		err := errors.New("test error")

		// This is a non-blocking call
		p.Shutdown(err)

		// No way to directly test the effect, but we can verify it doesn't panic
	})
}

// TestPal_Services tests the Services method
func TestPal_Services(t *testing.T) {
	t.Parallel()

	t.Run("returns all services", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		p := pal.New(service)

		assert.NoError(t, p.Init(t.Context()))

		services := p.Services()

		assert.Equal(t, []core.Service{service}, services)
	})

	t.Run("returns empty slice for no services", func(t *testing.T) {
		t.Parallel()

		p := pal.New()

		services := p.Services()

		assert.Empty(t, services)
	})
}

// TestPal_Invoke tests the Invoke method
func TestPal_Invoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes a service successfully", func(t *testing.T) {
		t.Parallel()

		p := pal.New(
			pal.Provide[TestInterface, TestStruct](),
		)

		assert.NoError(t, p.Init(t.Context()))

		instance, err := p.Invoke(t.Context(), "pal_test.TestInterface")
		assert.NoError(t, err)
		assert.Implements(t, (*TestInterface)(nil), instance)
	})

	t.Run("returns error when service not found", func(t *testing.T) {
		t.Parallel()

		p := pal.New()

		_, err := p.Invoke(t.Context(), "nonexistent")

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceNotFound)
	})
}

// TestPal_Invoke tests the Run method
func TestPal_Run(t *testing.T) {
	t.Parallel()

	t.Run("exists immediately when no runners given", func(t *testing.T) {
		t.Parallel()

		err := pal.New().
			InitTimeout(3*time.Second).
			HealthCheckTimeout(1*time.Second).
			ShutdownTimeout(3*time.Second).
			Run(t.Context(), syscall.SIGINT)

		assert.NoError(t, err)
	})

	t.Run("exists after runners exist", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[RunnerInterface, RunnerStruct]()

		err := pal.New(
			service,
		).
			InitTimeout(3*time.Second).
			HealthCheckTimeout(1*time.Second).
			ShutdownTimeout(3*time.Second).
			Run(t.Context(), syscall.SIGINT)

		require.NoError(t, err)

		runner, err := service.Instance(t.Context())
		assert.NoError(t, err)
		assert.NotNil(t, runner)

		assert.True(t, runner.(*RunnerStruct).RunCalled)
	})

	t.Run("errors during init - services are gracefully shut down", func(t *testing.T) {
		t.Parallel()

		// Create a test state tracker
		tracker := NewTestStateTracker()

		// Create a context with the test state tracker
		ctx := WithTestState(t.Context(), tracker)

		// Create a service that will be initialized successfully
		shutdownService := pal.Provide[ShutdownTrackingInterface, ShutdownTrackingStruct]()

		// Create a service that will fail during initialization
		failingService := pal.Provide[FailingInitInterface, FailingInitStruct]()

		// Create a runner that should not be started
		runnerService := pal.Provide[RunnerInterface, RunnerStruct]()

		// Run the application - this should fail because failingService fails to initialize
		err := pal.New(
			shutdownService,
			failingService,
			runnerService,
		).
			InitTimeout(3*time.Second).
			HealthCheckTimeout(1*time.Second).
			ShutdownTimeout(3*time.Second).
			Run(ctx, syscall.SIGINT)

		// Verify that Run returns an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "init error")

		// Verify that the shutdown service was shut down
		assert.True(t, tracker.ShutdownTrackerCalled())

		// Verify that the runner was not started
		assert.False(t, tracker.RunnerCalled())
	})

	t.Run("runners returning errors - services are gracefully shut down", func(t *testing.T) {
		t.Parallel()

		// Create a test state tracker
		tracker := NewTestStateTracker()

		// Create a context with the test state tracker
		ctx := WithTestState(t.Context(), tracker)

		// Create a service that will track if it was shut down
		shutdownService := pal.Provide[ShutdownTrackingInterface, ShutdownTrackingStruct]()

		// Create a runner that will return an error
		errorRunnerService := pal.Provide[ErrorRunnerInterface, ErrorRunnerStruct]()

		// Create a normal runner
		runnerService := pal.Provide[RunnerInterface, RunnerStruct]()

		// Run the application - this should fail because errorRunnerService returns an error
		err := pal.New(
			shutdownService,
			errorRunnerService,
			runnerService,
		).
			InitTimeout(3*time.Second).
			HealthCheckTimeout(1*time.Second).
			ShutdownTimeout(3*time.Second).
			Run(ctx, syscall.SIGINT)

		// Verify that Run returns an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "run error")

		// Verify that the error runner's Run was called
		assert.True(t, tracker.ErrorRunnerCalled())

		// Verify that the shutdown service was shut down
		assert.True(t, tracker.ShutdownTrackerCalled())

		// Verify that the normal runner was started
		assert.True(t, tracker.RunnerCalled())
	})
}

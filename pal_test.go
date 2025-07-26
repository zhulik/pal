package pal_test

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zhulik/pal"
)

// TestPal_New tests the New function
func Test_New(t *testing.T) {
	t.Parallel()

	t.Run("creates a new Pal instance with no services", func(t *testing.T) {
		t.Parallel()

		p := newPal()

		assert.NotNil(t, p)
		assert.Contains(t, p.Services(), "*pal.Pal")
	})

	t.Run("creates a new Pal instance with services", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.Provide(&TestServiceStruct{}).
				ToInit(func(ctx context.Context, service *TestServiceStruct) error {
					eventuallyAssertExpectations(t, service)
					service.On("Init", ctx).Return(nil)

					return nil
				}),
		)

		p = newPal(pal.ProvidePal(p))

		assert.NoError(t, p.Init(t.Context()))
		assert.Contains(t, p.Services(), "*pal_test.TestServiceStruct")
	})

	t.Run("correctly initializes service lists", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.ProvideList(
				pal.ProvideList(
					pal.Provide(&TestServiceStruct{}).
						ToInit(func(ctx context.Context, service *TestServiceStruct) error {
							eventuallyAssertExpectations(t, service)
							service.On("Init", ctx).Return(nil)

							return nil
						}),
				),
			),
		)

		require.NoError(t, p.Init(t.Context()))
	})
}

// TestPal_FromContext tests the FromContext function
func Test_FromContext(t *testing.T) {
	t.Parallel()

	t.Run("retrieves Pal from context", func(t *testing.T) {
		t.Parallel()

		p := newPal()
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

		p := newPal()
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

		p := newPal()
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

		p := newPal()
		timeout := 5 * time.Second

		result := p.ShutdownTimeout(timeout)

		assert.Same(t, p, result) // Method should return the Pal instance for chaining
	})
}

// TestPal_HealthCheck tests the HealthCheck method
func TestPal_HealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("performs health check on all services", func(t *testing.T) {
		t.Parallel()

		// Create a service that implements HealthChecker
		service := pal.Provide(&TestServiceStruct{})
		p := newPal(service)

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

		p := newPal()

		// This is a non-blocking call
		p.Shutdown()

		// No way to directly test the effect, but we can verify it doesn't panic
	})

	t.Run("schedules shutdown with errors", func(t *testing.T) {
		t.Parallel()

		p := newPal()

		// This is a non-blocking call
		p.Shutdown(errTest)

		// No way to directly test the effect, but we can verify it doesn't panic
	})
}

// TestPal_Services tests the Services method
func TestPal_Services(t *testing.T) {
	t.Parallel()

	t.Run("returns all services", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide(&TestServiceStruct{}).
			ToInit(func(ctx context.Context, service *TestServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Init", ctx).Return(nil)

				return nil
			})

		p := newPal(service)

		assert.NoError(t, p.Init(t.Context()))

		services := p.Services()

		assert.Contains(t, services, "*pal_test.TestServiceStruct")
		assert.Contains(t, services, "*pal.Pal")
	})

	t.Run("returns a slice with only pal for no services", func(t *testing.T) {
		t.Parallel()

		p := newPal()
		assert.NoError(t, p.Init(t.Context()))

		services := p.Services()

		assert.Contains(t, services, "*pal.Pal")
	})
}

// TestPal_Invoke tests the Invoke method
func TestPal_Invoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes a service successfully", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.Provide(&TestServiceStruct{}).
				ToInit(func(ctx context.Context, service *TestServiceStruct) error {
					eventuallyAssertExpectations(t, service)
					service.On("Init", ctx).Return(nil)

					return nil
				}),
		)

		assert.NoError(t, p.Init(t.Context()))

		instance, err := p.Invoke(t.Context(), "*pal_test.TestServiceStruct")
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("returns error when service not found", func(t *testing.T) {
		t.Parallel()

		p := newPal()

		_, err := p.Invoke(t.Context(), "nonexistent")

		assert.ErrorIs(t, err, pal.ErrServiceNotFound)
	})
}

// TestPal_Invoke tests the Run method
func TestPal_Run(t *testing.T) {
	t.Parallel()

	t.Run("exists immediately when no runners given", func(t *testing.T) {
		t.Parallel()

		err := newPal().
			InitTimeout(3*time.Second).
			HealthCheckTimeout(1*time.Second).
			ShutdownTimeout(3*time.Second).
			Run(t.Context(), syscall.SIGINT)

		assert.NoError(t, err)
	})

	t.Run("exists after runners exist", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide(&RunnerServiceStruct{}).
			ToInit(func(_ context.Context, service *RunnerServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Run", mock.Anything).Return(nil)

				return nil
			})

		err := newPal(
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

		i, _ := service.Instance(t.Context())

		m := i.(*RunnerServiceStruct)
		m.AssertExpectations(t)

		// assert.True(t, runner.(*RunnerServiceStruct).RunCalled)
	})

	t.Run("errors during init - services are gracefully shut down", func(t *testing.T) {
		t.Parallel()

		// Create a service that will be initialized successfully
		shutdownService := pal.Provide(&TestServiceStruct{}).
			ToInit(func(ctx context.Context, service *TestServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Init", ctx).Return(nil)
				service.On("Shutdown", ctx).Return(nil)

				return nil
			})

		// Create a service that will fail during initialization
		failingService := pal.Provide(&TestServiceStruct{}).
			ToInit(func(ctx context.Context, service *TestServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Init", ctx).Return(errTest)

				return nil
			})

		// Create a runner that should not be started
		runnerService := pal.Provide(&RunnerServiceStruct{})

		// Run the application - this should fail because failingService fails to initialize
		err := newPal(
			shutdownService,
			failingService,
			runnerService,
		).
			InitTimeout(3*time.Second).
			HealthCheckTimeout(1*time.Second).
			ShutdownTimeout(3*time.Second).
			Run(t.Context(), syscall.SIGINT)

		// Verify that Run returns an error
		require.Error(t, err)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("runners returning errors - services are gracefully shut down", func(t *testing.T) {
		t.Parallel()

		// Create a service that will track if it was shut down
		shutdownService := pal.Provide(&TestServiceStruct{}).
			ToInit(func(ctx context.Context, service *TestServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Init", ctx).Return(nil)
				service.On("Shutdown", mock.Anything).Return(nil)

				return nil
			})

		// for a different name in the container
		type errorRunnerInterface = TestServiceInterface
		// Create a runner that will return an error
		errorRunnerService := pal.Provide[errorRunnerInterface](&RunnerServiceStruct{}).
			ToInit(func(_ context.Context, service errorRunnerInterface) error {
				eventuallyAssertExpectations(t, service)
				service.(*RunnerServiceStruct).On("Run", mock.Anything).Return(errTest)

				return nil
			})

		// Create a normal runner
		runnerService := pal.Provide(&RunnerServiceStruct{}).
			ToInit(func(_ context.Context, service *RunnerServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Run", mock.Anything).Return(nil)

				return nil
			})

		// Run the application - this should fail because errorRunnerService returns an error
		err := newPal(
			shutdownService,
			errorRunnerService,
			runnerService,
		).
			InitTimeout(3*time.Second).
			HealthCheckTimeout(1*time.Second).
			ShutdownTimeout(3*time.Second).
			Run(t.Context(), syscall.SIGINT)

		// Verify that Run returns an error
		assert.ErrorIs(t, err, errTest)
	})
}

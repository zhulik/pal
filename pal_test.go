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
		assert.Contains(t, p.Services(), "*github.com/zhulik/pal.Pal")
	})

	t.Run("creates a new Pal instance with services", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
				s := &TestServiceStruct{}
				eventuallyAssertExpectations(t, s)
				s.On("Init", ctx).Return(nil)
				return s, nil
			}),
		)

		p = newPal(pal.ProvidePal(p))

		assert.NoError(t, p.Init(t.Context()))
		assert.Contains(t, p.Services(), "*github.com/zhulik/pal_test.TestServiceStruct")
	})

	t.Run("correctly initializes service lists", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.ProvideList(
				pal.ProvideList(
					pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
						s := &TestServiceStruct{}
						eventuallyAssertExpectations(t, s)
						s.On("Init", ctx).Return(nil)
						return s, nil
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

// TestPal_Services tests the Services method
func TestPal_Services(t *testing.T) {
	t.Parallel()

	t.Run("returns all services", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
			s := &TestServiceStruct{}
			eventuallyAssertExpectations(t, s)
			s.On("Init", ctx).Return(nil)
			return s, nil
		})

		p := newPal(service)

		assert.NoError(t, p.Init(t.Context()))

		services := p.Services()

		assert.Contains(t, services, "*github.com/zhulik/pal_test.TestServiceStruct")
		assert.Contains(t, services, "*github.com/zhulik/pal.Pal")
	})

	t.Run("returns a slice with only pal for no services", func(t *testing.T) {
		t.Parallel()

		p := newPal()
		assert.NoError(t, p.Init(t.Context()))

		services := p.Services()

		assert.Contains(t, services, "*github.com/zhulik/pal.Pal")
	})
}

// TestPal_Invoke tests the Invoke method
func TestPal_Invoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes a service successfully", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
				s := &TestServiceStruct{}
				eventuallyAssertExpectations(t, s)
				s.On("Init", ctx).Return(nil)
				return s, nil
			}),
		)

		assert.NoError(t, p.Init(t.Context()))

		instance, err := p.Invoke(t.Context(), "*github.com/zhulik/pal_test.TestServiceStruct")
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
		t.Skip("TODO: if not runners started, this will block forever, we should exit immediately if there are no runners")
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

		service := pal.ProvideFn(func(_ context.Context) (*RunnerServiceStruct, error) {
			s := &RunnerServiceStruct{}
			eventuallyAssertExpectations(t, s)
			s.On("Run", mock.Anything).Return(nil)
			return s, nil
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
	})

	t.Run("errors during init - services are gracefully shut down", func(t *testing.T) {
		t.Skip("TODO: this test is not working as expected")
		t.Parallel()

		// Create a service that will be initialized successfully
		shutdownService := pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
			s := &TestServiceStruct{}
			eventuallyAssertExpectations(t, s)
			s.On("Init", ctx).Return(nil)
			s.On("Shutdown", ctx).Return(nil)
			return s, nil
		})

		// Create a service that will fail during initialization
		failingService := pal.Provide(&TestServiceStruct{}).
			ToInit(func(_ context.Context, _ *TestServiceStruct, _ *pal.Pal) error {
				return errTest
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
		t.Skip("TODO: this test is not working as expected")
		t.Parallel()

		// Create a service that will track if it was shut down
		shutdownService := pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
			s := &TestServiceStruct{}
			eventuallyAssertExpectations(t, s)
			s.On("Init", ctx).Return(nil)
			s.On("Shutdown", mock.Anything).Return(nil)
			return s, nil
		})

		// for a different name in the container
		type errorRunnerInterface = TestServiceInterface
		// Create a runner that will return an error
		errorRunnerService := pal.ProvideFn(func(_ context.Context) (errorRunnerInterface, error) {
			s := &RunnerServiceStruct{}
			eventuallyAssertExpectations(t, s)
			s.On("Run", mock.Anything).Return(errTest)
			return s, nil
		})

		// Create a normal runner
		runnerService := pal.ProvideFn(func(_ context.Context) (*RunnerServiceStruct, error) {
			s := &RunnerServiceStruct{}
			eventuallyAssertExpectations(t, s)
			s.On("Run", mock.Anything).Return(nil)
			return s, nil
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

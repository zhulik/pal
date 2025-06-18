package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
)

// Test interfaces and implementations are defined in common_test.go

// TestProvide tests the Provide function
func TestProvide(t *testing.T) {
	t.Parallel()

	t.Run("creates a singleton service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]()

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.TestServiceInterface", service.Name())
	})

	t.Run("detects runner services", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[RunnerServiceInterface, RunnerServiceStruct]().
			BeforeInit(func(ctx context.Context, service *RunnerServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Run", ctx).Return(nil)

				return nil
			})

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.RunnerServiceInterface", service.Name())
	})
}

// TestProvideFactory tests the ProvideFactory function
func TestProvideFactory(t *testing.T) {
	t.Parallel()

	t.Run("creates a factory service", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory[TestServiceInterface, TestServiceStruct]()

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.TestServiceInterface", service.Name())
	})
}

// TestProvideFn tests the ProvideFn function
func TestProvideFn(t *testing.T) {
	t.Parallel()

	t.Run("creates a singleton service with a function", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFn(func(ctx context.Context) (TestServiceInterface, error) {
			s := &TestServiceStruct{}
			s.On("Init", ctx).Return(nil)
			return s, nil
		})

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.TestServiceInterface", service.Name())
	})
}

// TestProvideFnFactory tests the ProvideFnFactory function
func TestProvideFnFactory(t *testing.T) {
	t.Parallel()

	t.Run("creates a factory service with a function", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFnFactory[TestServiceInterface](func(_ context.Context) (TestServiceInterface, error) {
			return &TestServiceStruct{}, nil
		})

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.TestServiceInterface", service.Name())
	})
}

// TestProvideConst tests the ProvideConst function
func TestProvideConst(t *testing.T) {
	t.Parallel()

	t.Run("creates a const service", func(t *testing.T) {
		t.Parallel()

		s := &TestServiceStruct{}
		service := pal.ProvideConst[TestServiceInterface](s)

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.TestServiceInterface", service.Name())

		// Verify that the instance is the same
		instance, err := service.Instance(context.Background())
		assert.NoError(t, err)
		assert.Same(t, s, instance)
	})
}

// TestInvoke tests the Invoke function
func TestInvoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes a service successfully", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.Provide[TestServiceInterface, TestServiceStruct]().
				BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
					eventuallyAssertExpectations(t, service)
					service.On("Init", ctx).Return(nil)

					return nil
				}),
		)

		require.NoError(t, p.Init(t.Context()))

		instance, err := pal.Invoke[TestServiceInterface](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("returns error when service not found", func(t *testing.T) {
		t.Parallel()

		// Create an empty Pal instance
		p := newPal()

		// Try to invoke a non-existent service
		_, err := pal.Invoke[TestServiceInterface](t.Context(), p)

		assert.ErrorIs(t, err, pal.ErrServiceNotFound)
	})
}

// TestBuild tests the Build function
func TestBuild(t *testing.T) {
	t.Parallel()

	t.Run("injects dependencies successfully", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.Provide[TestServiceInterface, TestServiceStruct]().
				BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
					eventuallyAssertExpectations(t, service)
					service.On("Init", ctx).Return(nil)

					return nil
				}),
		)

		require.NoError(t, p.Init(t.Context()))

		type DependentStruct struct {
			Dependency TestServiceInterface
		}

		instance, err := pal.Build[DependentStruct](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.NotNil(t, instance.Dependency)
	})

	t.Run("ignores missing dependencies", func(t *testing.T) {
		t.Parallel()

		// Create an empty Pal instance
		p := newPal()

		// Try to inject dependencies with no services registered
		_, err := pal.Build[DependentStruct](t.Context(), p)

		assert.NoError(t, err)
	})

	t.Run("skips non-interface fields", func(t *testing.T) {
		t.Parallel()

		type StructWithNonInterfaceField struct {
			NonInterface string
		}

		// Create an empty Pal instance
		p := newPal()

		// Build dependencies into a struct with no interface fields
		result, err := pal.Build[StructWithNonInterfaceField](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.NonInterface) // Default value is empty string
	})

	t.Run("skips unexported fields", func(t *testing.T) {
		t.Parallel()

		type StructWithUnexportedField struct {
			dependency TestServiceInterface
		}

		// Create a Pal instance with our test service
		p := newPal(pal.Provide[TestServiceInterface, TestServiceStruct]())

		// No need to initialize Pal for this test

		// Build dependencies into a struct with unexported fields
		result, err := pal.Build[StructWithUnexportedField](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.dependency) // Field is unexported, so it's not set
	})
}

// TestInjectInto tests the InjectInto function
func TestInjectInto(t *testing.T) {
	t.Parallel()

	t.Run("injects dependencies successfully", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.Provide[TestServiceInterface, TestServiceStruct]().
				BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
					eventuallyAssertExpectations(t, service)
					service.On("Init", ctx).Return(nil)

					return nil
				}),
		)

		require.NoError(t, p.Init(t.Context()))

		// Create a struct instance to inject dependencies into
		instance := &DependentStruct{}

		// Inject dependencies
		err := pal.InjectInto(t.Context(), p, instance)

		assert.NoError(t, err)
		assert.NotNil(t, instance.Dependency)
	})

	t.Run("ignores missing dependencies", func(t *testing.T) {
		t.Parallel()

		// Create an empty Pal instance
		p := newPal()

		// Create a struct instance to inject dependencies into
		instance := &DependentStruct{}

		// Try to inject dependencies with no services registered
		err := pal.InjectInto(t.Context(), p, instance)

		assert.NoError(t, err)
		assert.Nil(t, instance.Dependency) // Dependency should remain nil
	})

	t.Run("skips non-interface fields", func(t *testing.T) {
		t.Parallel()

		type StructWithNonInterfaceField struct {
			NonInterface string
		}

		// Create an empty Pal instance
		p := newPal()

		// Create a struct instance to inject dependencies into
		instance := &StructWithNonInterfaceField{NonInterface: "original value"}

		// Inject dependencies
		err := pal.InjectInto(t.Context(), p, instance)

		assert.NoError(t, err)
		assert.Equal(t, "original value", instance.NonInterface) // Value should remain unchanged
	})

	t.Run("skips unexported fields", func(t *testing.T) {
		t.Parallel()

		type StructWithUnexportedField struct {
			dependency TestServiceInterface
		}

		// Create a Pal instance with our test service
		p := newPal(pal.Provide[TestServiceInterface, TestServiceStruct]())

		// Create a struct instance to inject dependencies into
		instance := &StructWithUnexportedField{}

		// Inject dependencies
		err := pal.InjectInto(t.Context(), p, instance)

		assert.NoError(t, err)
		assert.Nil(t, instance.dependency) // Field is unexported, so it's not set
	})

	t.Run("returns error when service invocation fails", func(t *testing.T) {
		t.Parallel()

		// Create a struct instance to inject dependencies into
		instance := &DependentStruct{}

		// Create a mock invoker that returns an error
		mockInvoker := &MockInvoker{}
		mockInvoker.On("InjectInto", mock.Anything, instance).Return(errTest)

		// Inject dependencies
		err := pal.InjectInto(t.Context(), mockInvoker, instance)

		assert.ErrorIs(t, err, errTest)
		assert.Nil(t, instance.Dependency) // Dependency should remain nil
	})
}

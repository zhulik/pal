package pal_test

import (
	"context"
	"testing"

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
		assert.True(t, service.IsSingleton())
		assert.False(t, service.IsRunner())
	})

	t.Run("detects runner services", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[RunnerServiceInterface, RunnerServiceStruct]().BeforeInit(func(ctx context.Context, service *RunnerServiceStruct) error {
			eventuallyAssertExpectations(t, service)
			service.On("Run", ctx).Return(nil)

			return nil
		})

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.RunnerServiceInterface", service.Name())
		assert.True(t, service.IsSingleton())
		assert.True(t, service.IsRunner())
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
		assert.False(t, service.IsSingleton())
		assert.False(t, service.IsRunner())
	})
}

// TestInvoke tests the Invoke function
func TestInvoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes a service successfully", func(t *testing.T) {
		t.Parallel()

		p := pal.New(
			pal.Provide[TestServiceInterface, TestServiceStruct]().BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
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
		p := pal.New()

		// Try to invoke a non-existent service
		_, err := pal.Invoke[TestServiceInterface](t.Context(), p)

		assert.ErrorIs(t, err, pal.ErrServiceNotFound)
	})
}

// TestInject tests the Inject function
func TestInject(t *testing.T) {
	t.Parallel()

	t.Run("injects dependencies successfully", func(t *testing.T) {
		t.Parallel()

		p := pal.New(
			pal.Provide[TestServiceInterface, TestServiceStruct]().BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
				eventuallyAssertExpectations(t, service)
				service.On("Init", ctx).Return(nil)

				return nil
			}),
		)

		require.NoError(t, p.Init(t.Context()))

		type DependentStruct struct {
			Dependency TestServiceInterface
		}

		instance, err := pal.Inject[DependentStruct](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.NotNil(t, instance.Dependency)
	})

	t.Run("returns error when dependency not found", func(t *testing.T) {
		t.Parallel()

		// Create an empty Pal instance
		p := pal.New()

		// Try to inject dependencies with no services registered
		_, err := pal.Inject[DependentStruct](t.Context(), p)

		assert.ErrorIs(t, err, pal.ErrServiceNotFound)
	})

	t.Run("skips non-interface fields", func(t *testing.T) {
		t.Parallel()

		type StructWithNonInterfaceField struct {
			NonInterface string
		}

		// Create an empty Pal instance
		p := pal.New()

		// Inject dependencies into a struct with no interface fields
		result, err := pal.Inject[StructWithNonInterfaceField](t.Context(), p)

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
		p := pal.New(pal.Provide[TestServiceInterface, TestServiceStruct]())

		// No need to initialize Pal for this test

		// Inject dependencies into a struct with unexported fields
		result, err := pal.Inject[StructWithUnexportedField](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.dependency) // Field is unexported, so it's not set
	})
}

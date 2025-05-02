package pal_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
	"github.com/zhulik/pal/pkg/core"
)

// Test interfaces and implementations are defined in common_test.go

// TestProvide tests the Provide function
func TestProvide(t *testing.T) {
	t.Parallel()

	t.Run("creates a singleton service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.TestInterface", service.Name())
		assert.True(t, service.IsSingleton())
		assert.False(t, service.IsRunner())
	})

	t.Run("detects runner services", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[RunnerInterface, RunnerStruct]()

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.RunnerInterface", service.Name())
		assert.True(t, service.IsSingleton())
		assert.True(t, service.IsRunner())
	})
}

// TestProvideFactory tests the ProvideFactory function
func TestProvideFactory(t *testing.T) {
	t.Parallel()

	t.Run("creates a factory service", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory[TestInterface, TestStruct]()

		assert.NotNil(t, service)
		assert.Equal(t, "pal_test.TestInterface", service.Name())
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
			pal.Provide[TestInterface, TestStruct](),
		)

		require.NoError(t, p.Init(t.Context()))

		instance, err := pal.Invoke[TestInterface](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("returns error when service not found", func(t *testing.T) {
		t.Parallel()

		// Create an empty Pal instance
		p := pal.New()

		// Try to invoke a non-existent service
		_, err := pal.Invoke[TestInterface](t.Context(), p)

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceNotFound)
	})
}

// TestInject tests the Inject function
func TestInject(t *testing.T) {
	t.Parallel()

	t.Run("injects dependencies successfully", func(t *testing.T) {
		t.Parallel()

		p := pal.New(
			pal.Provide[TestInterface, TestStruct](),
		)

		require.NoError(t, p.Init(t.Context()))

		type DependentStruct struct {
			Dependency TestInterface
		}

		instance, err := pal.Inject[DependentStruct](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("returns error when dependency not found", func(t *testing.T) {
		t.Parallel()

		// Create an empty Pal instance
		p := pal.New()

		// Try to inject dependencies with no services registered
		_, err := pal.Inject[DependentStruct](t.Context(), p)

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceNotFound)
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
			dependency TestInterface
		}

		// Create a Pal instance with our test service
		p := pal.New(pal.Provide[TestInterface, TestStruct]())

		// No need to initialize Pal for this test

		// Inject dependencies into a struct with unexported fields
		result, err := pal.Inject[StructWithUnexportedField](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.dependency) // Field is unexported, so it's not set
	})
}

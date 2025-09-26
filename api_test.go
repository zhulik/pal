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

		service := pal.Provide(NewMockTestServiceStruct(t))

		assert.NotNil(t, service)
		assert.Equal(t, "*github.com/zhulik/pal_test.TestServiceStruct", service.Name())
	})

	t.Run("detects runner services", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide(NewMockRunnerServiceStruct(t)).
			ToInit(func(ctx context.Context, service *RunnerServiceStruct, _ *pal.Pal) error {
				service.MockRunner.EXPECT().Run(ctx).Return(nil)

				return nil
			})

		assert.NotNil(t, service)
		assert.Equal(t, "*github.com/zhulik/pal_test.RunnerServiceStruct", service.Name())
	})

	t.Run("makes sure the argument is a pointer to struct", func(t *testing.T) {
		t.Parallel()

		require.PanicsWithValue(t, "Argument must be a non-nil pointer to a struct, got func()", func() {
			pal.Provide(func() {})
		})
	})
}

// TestProvideFn tests the ProvideFn function
func TestProvideFn(t *testing.T) {
	t.Parallel()

	t.Run("creates a singleton service with a function", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFn(func(ctx context.Context) (TestServiceInterface, error) {
			s := NewMockTestServiceStruct(t)
			s.MockIniter.EXPECT().Init(ctx).Return(nil)
			return s, nil
		})

		assert.NotNil(t, service)
		assert.Equal(t, "github.com/zhulik/pal_test.TestServiceInterface", service.Name())
	})
}

// TestProvideFactory0 tests the ProvideFactory0 function
func TestProvideFactory0(t *testing.T) {
	t.Parallel()

	t.Run("creates a factory service with a function", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0(func(_ context.Context) (TestServiceInterface, error) {
			return NewMockTestServiceStruct(t), nil
		})

		assert.NotNil(t, service)
		assert.Equal(t, "github.com/zhulik/pal_test.TestServiceInterface", service.Name())
	})
}

// TestProvideConst tests the ProvideConst function
func TestProvideConst(t *testing.T) {
	t.Parallel()

	t.Run("creates a const service", func(t *testing.T) {
		t.Parallel()

		s := NewMockTestServiceStruct(t)
		service := pal.Provide(s)

		assert.NotNil(t, service)
		assert.Equal(t, "*github.com/zhulik/pal_test.TestServiceStruct", service.Name())

		// Verify that the instance is the same
		instance, err := service.Instance(t.Context())
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
			pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
				s := NewMockTestServiceStruct(t)
				s.MockIniter.EXPECT().Init(ctx).Return(nil)
				return s, nil
			}),
		)

		require.NoError(t, p.Init(t.Context()))

		instance, err := pal.Invoke[*TestServiceStruct](t.Context(), p)

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

func TestInvokeAs(t *testing.T) {
	t.Parallel()

	t.Run("invokes a service successfully", func(t *testing.T) {
		t.Parallel()

		p := newPal(pal.Provide[TestServiceInterface](NewMockTestServiceStruct(t)))

		instance, err := pal.InvokeAs[TestServiceInterface, TestServiceStruct](t.Context(), p)

		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("returns error when service cannot be cast to the expected type", func(t *testing.T) {
		t.Parallel()

		p := newPal(pal.Provide[TestServiceInterface](NewMockTestServiceStruct(t)))

		_, err := pal.InvokeAs[TestServiceInterface, string](t.Context(), p)

		assert.ErrorIs(t, err, pal.ErrServiceInvalidCast)
	})
}

// TestBuild tests the Build function
func TestBuild(t *testing.T) {
	t.Parallel()

	t.Run("injects dependencies successfully", func(t *testing.T) {
		t.Parallel()

		p := newPal(
			pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
				s := NewMockTestServiceStruct(t)
				s.MockIniter.EXPECT().Init(ctx).Return(nil)
				return s, nil
			}),
		)

		require.NoError(t, p.Init(t.Context()))

		type DependentStruct struct {
			Dependency *TestServiceStruct
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
		p := newPal(pal.Provide(NewMockTestServiceStruct(t)))

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
			pal.ProvideFn(func(ctx context.Context) (*TestServiceStruct, error) {
				s := NewMockTestServiceStruct(t)
				s.MockIniter.EXPECT().Init(ctx).Return(nil)
				return s, nil
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
			dependency *TestServiceStruct
		}

		// Create a Pal instance with our test service
		p := newPal(pal.Provide(NewMockTestServiceStruct(t)))

		// Create a struct instance to inject dependencies into
		instance := &StructWithUnexportedField{}

		// Inject dependencies
		err := pal.InjectInto(t.Context(), p, instance)

		assert.NoError(t, err)
		assert.Nil(t, instance.dependency) // Field is unexported, so it's not set
	})

	t.Run("skips fields with skip tag", func(t *testing.T) {
		t.Parallel()

		type StructWithSkipField struct {
			Dependency *TestServiceStruct `pal:"skip"`
		}

		p := newPal(pal.Provide(NewMockTestServiceStruct(t)))

		instance := &StructWithSkipField{}

		err := pal.InjectInto(t.Context(), p, instance)

		assert.NoError(t, err)
		assert.Nil(t, instance.Dependency) // Field is skipped, so it's not set
	})

	t.Run("returns error when service invocation fails", func(t *testing.T) {
		t.Parallel()

		// Create a struct instance to inject dependencies into
		instance := &DependentStruct{}

		// Create a mock invoker that returns an error
		mockInvoker := NewMockInvoker(t)
		mockInvoker.EXPECT().InjectInto(mock.Anything, instance).Return(errTest)

		// Inject dependencies
		err := pal.InjectInto(t.Context(), mockInvoker, instance)

		assert.ErrorIs(t, err, errTest)
		assert.Nil(t, instance.Dependency) // Dependency should remain nil
	})
}

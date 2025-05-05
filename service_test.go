package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
)

// TestService_Name tests the Name method of the service struct
func TestService_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns correct name for interface type", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]()

		assert.Equal(t, "pal_test.TestServiceInterface", service.Name())
	})
}

// TestService_IsSingleton tests the IsSingleton method of the service struct
func TestService_IsSingleton(t *testing.T) {
	t.Parallel()

	t.Run("returns true for singleton service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]()

		assert.True(t, service.IsSingleton())
	})

	t.Run("returns false for factory service", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory[TestServiceInterface, TestServiceStruct]()

		assert.False(t, service.IsSingleton())
	})
}

// TestService_IsRunner tests the IsRunner method of the service struct
func TestService_IsRunner(t *testing.T) {
	t.Parallel()

	t.Run("returns true for runner service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[RunnerServiceInterface, RunnerServiceStruct]()

		assert.True(t, service.IsRunner())
	})

	t.Run("returns false for non-runner service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]()

		assert.False(t, service.IsRunner())
	})
}

// TestService_Make tests the Make method of the service struct
func TestService_Make(t *testing.T) {
	t.Parallel()

	t.Run("returns empty instance of service type", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]().BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
			eventuallyAssertExpectations(t, service)
			service.On("Init", ctx).Return(nil)

			return nil
		})

		assert.Implements(t, (*TestServiceInterface)(nil), service.Make())
	})
}

// TestService_Validate tests the Validate method of the service struct
func TestService_Validate(t *testing.T) {
	t.Parallel()

	t.Run("validates service successfully", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]()

		assert.NoError(t, service.Validate(t.Context()))
	})

	t.Run("returns error when I is not an interface", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceStruct, TestServiceStruct]()

		err := service.Validate(t.Context())

		assert.ErrorIs(t, err, pal.ErrServiceInvalid)
	})

	t.Run("returns error when S is not a struct", func(t *testing.T) {
		t.Parallel()

		type NotAStruct string
		service := pal.Provide[TestServiceInterface, NotAStruct]()

		err := service.Validate(t.Context())

		assert.ErrorIs(t, err, pal.ErrServiceInvalid)
	})

	t.Run("returns error when S does not implement I", func(t *testing.T) {
		t.Parallel()

		type NonImplementingStruct struct{}
		service := pal.Provide[TestServiceInterface, NonImplementingStruct]()

		err := service.Validate(t.Context())

		assert.ErrorIs(t, err, pal.ErrServiceInvalid)
	})
}

// TestService_Instance tests the Instance method of the service struct
func TestService_Instance(t *testing.T) {
	t.Parallel()

	t.Run("returns instance for singleton service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]().BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
			eventuallyAssertExpectations(t, service)
			service.On("Init", ctx).Return(nil)

			return nil
		})

		p := newPal(service)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance1, err := service.Instance(t.Context())

		assert.NoError(t, err)
		assert.NotNil(t, instance1)

		instance2, err := service.Instance(t.Context())

		assert.NoError(t, err)
		assert.NotNil(t, instance1)
		assert.Same(t, instance1, instance2)
	})

	t.Run("returns new instance for factory service", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory[TestServiceInterface, TestServiceStruct]().BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
			eventuallyAssertExpectations(t, service)
			service.On("Init", ctx).Return(nil)

			return nil
		})
		p := newPal(service)
		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		// First call to Instance should create a new instance
		instance, err := service.Instance(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, instance)

		// Second call should create another new instance
		instance2, err := service.Instance(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, instance2)
		assert.NotSame(t, instance, instance2)
	})
}

// TestService_BeforeInit tests the BeforeInit hook functionality
func TestService_BeforeInit(t *testing.T) {
	t.Parallel()

	t.Run("hook is called when set", func(t *testing.T) {
		t.Parallel()

		hook := func(ctx context.Context, service *TestServiceStruct) error {
			eventuallyAssertExpectations(t, service)
			service.On("Init", ctx).Return(nil)

			return nil
		}

		service := pal.Provide[TestServiceInterface, TestServiceStruct]().BeforeInit(hook)
		p := newPal(service)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance, err := service.Instance(t.Context())
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("no error when hook is not set", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestServiceInterface, TestServiceStruct]().BeforeInit(func(ctx context.Context, service *TestServiceStruct) error {
			eventuallyAssertExpectations(t, service)
			service.On("Init", ctx).Return(nil)

			return nil
		})
		p := newPal(service)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance, err := service.Instance(t.Context())
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("propagates error from hook", func(t *testing.T) {
		t.Parallel()

		hook := func(_ context.Context, _ *TestServiceStruct) error {
			return errTest
		}

		service := pal.Provide[TestServiceInterface, TestServiceStruct]().BeforeInit(hook)
		p := newPal(service)

		// The error should be propagated from the hook through Initialize to Init
		err := p.Init(t.Context())
		assert.ErrorIs(t, err, errTest)
	})
}

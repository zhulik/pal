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

		service := pal.Provide(&TestServiceStruct{})

		assert.Equal(t, "*pal_test.TestServiceStruct", service.Name())
	})
}

// TestService_Instance tests the Instance method of the service struct
func TestService_Instance(t *testing.T) {
	t.Parallel()

	t.Run("returns instance for singleton service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide(&RunnerServiceStruct{}).
			ToInit(func(ctx context.Context, service *RunnerServiceStruct, _ *pal.Pal) error {
				eventuallyAssertExpectations(t, service)

				service.On("Init", ctx).Return(nil)

				return nil
			})

		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance1, err := service.Instance(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, instance1)

		instance2, err := service.Instance(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, instance1)
		assert.Same(t, instance1, instance2)
	})
}

// TestService_ToInit tests the ToInit hook functionality
func TestService_ToInit(t *testing.T) {
	t.Parallel()

	t.Run("hook is called when set", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide(&TestServiceStruct{}).
			ToInit(func(ctx context.Context, service *TestServiceStruct, _ *pal.Pal) error {
				eventuallyAssertExpectations(t, service)
				service.On("Init", ctx).Return(nil)

				return nil
			})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance, err := service.Instance(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("no error when hook is not set", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide(&TestServiceStruct{}).
			ToInit(func(ctx context.Context, service *TestServiceStruct, _ *pal.Pal) error {
				eventuallyAssertExpectations(t, service)
				service.On("Init", ctx).Return(nil)

				return nil
			})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance, err := service.Instance(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("propagates error from hook", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide(&TestServiceStruct{}).
			ToInit(func(_ context.Context, _ *TestServiceStruct, _ *pal.Pal) error {
				return errTest
			})
		p := newPal(service)

		// The error should be propagated from the hook through Initialize to Init
		err := p.Init(t.Context())
		assert.ErrorIs(t, err, errTest)
	})
}

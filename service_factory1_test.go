package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
)

type Factory1Service struct {
	Name string
}

// TestService_Instance tests the Instance method of the service struct
func TestServiceFactory1_Instance(t *testing.T) {
	t.Parallel()

	t.Run("when called with correct arguments, returns a new instance built with given arguments", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1(func(_ context.Context, name string) (*Factory1Service, error) {
			return &Factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance1, err := service.Instance(ctx, "test")

		assert.NoError(t, err)
		assert.NotNil(t, instance1)

		assert.Equal(t, "test", instance1.(*Factory1Service).Name)

		instance2, err := service.Instance(ctx, "test2")

		assert.NoError(t, err)
		assert.NotNil(t, instance1)
		assert.Equal(t, "test2", instance2.(*Factory1Service).Name)

		assert.NotSame(t, instance1, instance2)
	})

	t.Run("when called with incorrect number of arguments, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1(func(_ context.Context, name string) (*Factory1Service, error) {
			return &Factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		_, err = service.Instance(ctx)

		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)
	})

	t.Run("when called with incorrect argument type, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1(func(_ context.Context, name string) (*Factory1Service, error) {
			return &Factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		_, err = service.Instance(ctx, 1)

		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
	})
}

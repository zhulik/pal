package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
)

type factory1Service struct {
	Name string
}

type serviceWithFactoryServiceDependency struct {
	Dependency *factory1Service
}

// TestService_Instance tests the Instance method of the service struct
func TestServiceFactory1_Instance(t *testing.T) {
	t.Parallel()

	t.Run("when called with correct arguments, returns a new instance built with given arguments", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1(func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance1, err := p.Invoke(ctx, service.Name(), "test")

		assert.NoError(t, err)
		assert.NotNil(t, instance1)

		assert.Equal(t, "test", instance1.(*factory1Service).Name)

		instance2, err := p.Invoke(ctx, service.Name(), "test2")

		assert.NoError(t, err)
		assert.NotNil(t, instance1)
		assert.Equal(t, "test2", instance2.(*factory1Service).Name)

		assert.NotSame(t, instance1, instance2)
	})

	t.Run("when called with incorrect number of arguments, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1(func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		_, err = p.Invoke(ctx, service.Name())

		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)
	})

	t.Run("when called with incorrect argument type, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1(func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		_, err = p.Invoke(ctx, service.Name(), 1)

		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
	})

	t.Run("when a service with a factory service dependency is invoked, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1(func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		err = p.InjectInto(ctx, &serviceWithFactoryServiceDependency{})

		assert.ErrorIs(t, err, pal.ErrFactoryServiceDependency)
	})
}

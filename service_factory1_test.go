package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

type factory1Service struct {
	Name string
}

type serviceWithFactoryServiceDependency struct {
	Dependency *factory1Service
}

type serviceWithFactoryFunctionDependency struct {
	CreateDependency     func(ctx context.Context, name string) (*factory1Service, error)
	MustCreateDependency func(ctx context.Context, name string) *factory1Service
}

// TestService_Instance tests the Instance method of the service struct
func TestServiceFactory1_Invocation(t *testing.T) {
	t.Parallel()

	t.Run("when invoked with correct arguments, returns a new instance built with given arguments", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := pal.WithPal(t.Context(), p)

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

	t.Run("when invoked with incorrect number of arguments, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := pal.WithPal(t.Context(), p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		_, err = p.Invoke(ctx, service.Name())

		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)
	})

	t.Run("when invoked with incorrect argument type, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := pal.WithPal(t.Context(), p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		_, err = p.Invoke(ctx, service.Name(), 1)

		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
	})

	t.Run("when a service with a factory service dependency is invoked, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := pal.WithPal(t.Context(), p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		err = p.InjectInto(ctx, &serviceWithFactoryServiceDependency{})

		assert.ErrorIs(t, err, pal.ErrFactoryServiceDependency)
	})

	t.Run("when invoked via injected factory function, returns a new instance built with given arguments", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := pal.WithPal(t.Context(), p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		serviceWithFactoryFn := &serviceWithFactoryFunctionDependency{}
		err = p.InjectInto(ctx, serviceWithFactoryFn)

		assert.NoError(t, err)

		dependency, err := serviceWithFactoryFn.CreateDependency(ctx, "test")

		assert.NoError(t, err)
		assert.Equal(t, "test", dependency.Name)
	})

	t.Run("when invoked via injected must factory function, returns a new instance built with given arguments", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, name string) (*factory1Service, error) {
			return &factory1Service{Name: name}, nil
		})
		p := newPal(service)

		ctx := pal.WithPal(t.Context(), p)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		serviceWithFactoryFn := &serviceWithFactoryFunctionDependency{}
		err = p.InjectInto(ctx, serviceWithFactoryFn)

		assert.NoError(t, err)

		dependency := serviceWithFactoryFn.MustCreateDependency(ctx, "test")

		assert.Equal(t, "test", dependency.Name)
	})
}

func TestServiceFactory1_instanceInitFailure(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory1[TestServiceInterface](func(ctx context.Context, _ string) (*TestServiceStruct, error) {
		s := NewMockTestServiceStruct(t)
		s.MockIniter.EXPECT().Init(ctx).Return(errTest)
		return s, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "x")
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory1_fnError(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory1[*factoryMultiLabel](func(_ context.Context, _ string) (*factoryMultiLabel, error) {
		return nil, errTest
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "x")
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory1_FactoryAndMustFactory(t *testing.T) {
	t.Parallel()

	t.Run("Factory propagates error from builder", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, _ string) (*factory1Service, error) {
			return nil, errTest
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		fn := service.Factory().(func(context.Context, string) (*factory1Service, error))
		_, err := fn(ctx, "x")
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("MustFactory panics when builder fails", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory1[*factory1Service](func(_ context.Context, _ string) (*factory1Service, error) {
			return nil, errTest
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		mustFn := service.MustFactory().(func(context.Context, string) *factory1Service)
		assert.PanicsWithValue(t, errTest, func() { mustFn(ctx, "x") })
	})
}

func TestServiceFactory1_firstArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory1[*factoryMultiLabel](func(_ context.Context, _ int) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "not-int")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

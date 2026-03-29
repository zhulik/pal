package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

func TestServiceFactory0_Invocation(t *testing.T) {
	t.Parallel()

	t.Run("when invoked with correct arguments count, returns a new instance each time", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0[*factoryMultiLabel](func(_ context.Context) (*factoryMultiLabel, error) {
			return &factoryMultiLabel{S0: "ok"}, nil
		})
		assert.Zero(t, service.Arguments())

		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		a, err := p.Invoke(ctx, service.Name())
		require.NoError(t, err)
		b, err := p.Invoke(ctx, service.Name())
		require.NoError(t, err)
		assert.NotSame(t, a, b)
		assert.Equal(t, "ok", a.(*factoryMultiLabel).S0)
	})

	t.Run("when invoked with extra arguments, returns invalid arguments count", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0[*factoryMultiLabel](func(_ context.Context) (*factoryMultiLabel, error) {
			return &factoryMultiLabel{}, nil
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		_, err := p.Invoke(ctx, service.Name(), "extra")
		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)
	})

	t.Run("when factory function fails, Invoke wraps ErrServiceInitFailed", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0[*factoryMultiLabel](func(_ context.Context) (*factoryMultiLabel, error) {
			return nil, errTest
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		_, err := p.Invoke(ctx, service.Name())
		assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("when instance init fails, Invoke wraps ErrServiceInitFailed", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0[TestServiceInterface](func(ctx context.Context) (*TestServiceStruct, error) {
			s := NewMockTestServiceStruct(t)
			s.MockIniter.EXPECT().Init(ctx).Return(errTest)
			return s, nil
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		_, err := p.Invoke(ctx, service.Name())
		assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("Factory closure returns error from underlying function", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0[*factoryMultiLabel](func(_ context.Context) (*factoryMultiLabel, error) {
			return nil, errTest
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		fn := service.Factory().(func(context.Context) (*factoryMultiLabel, error))
		_, err := fn(ctx)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("Factory closure returns instance on success", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0[*factoryMultiLabel](func(_ context.Context) (*factoryMultiLabel, error) {
			return &factoryMultiLabel{S0: "f0"}, nil
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		fn := service.Factory().(func(context.Context) (*factoryMultiLabel, error))
		got, err := fn(ctx)
		require.NoError(t, err)
		assert.Equal(t, "f0", got.S0)
	})

	t.Run("MustFactory panics when instance creation fails", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory0[*factoryMultiLabel](func(_ context.Context) (*factoryMultiLabel, error) {
			return nil, errTest
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		mustFn := service.MustFactory().(func(context.Context) *factoryMultiLabel)
		assert.PanicsWithValue(t, errTest, func() { mustFn(ctx) })
	})
}

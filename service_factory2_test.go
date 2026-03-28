package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

func TestServiceFactory2_Invocation(t *testing.T) {
	t.Parallel()

	t.Run("when invoked with correct arguments, returns a new instance", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory2[*factoryMultiLabel](func(_ context.Context, a string, b int) (*factoryMultiLabel, error) {
			return &factoryMultiLabel{S0: a, I0: b}, nil
		})
		assert.Equal(t, 2, service.Arguments())

		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		inst, err := p.Invoke(ctx, service.Name(), "x", 7)
		require.NoError(t, err)
		assert.Equal(t, "x", inst.(*factoryMultiLabel).S0)
		assert.Equal(t, 7, inst.(*factoryMultiLabel).I0)
	})

	t.Run("when invoked with wrong argument count, returns an error", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory2[*factoryMultiLabel](func(_ context.Context, _, _ string) (*factoryMultiLabel, error) {
			return &factoryMultiLabel{}, nil
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		_, err := p.Invoke(ctx, service.Name(), "only")
		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)
	})

	t.Run("when second argument has wrong type, returns invalid argument type", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory2[*factoryMultiLabel](func(_ context.Context, _ string, _ int) (*factoryMultiLabel, error) {
			return &factoryMultiLabel{}, nil
		})
		p := newPal(service)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(t.Context()))

		_, err := p.Invoke(ctx, service.Name(), "a", "b")
		assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
	})
}

func TestServiceFactory2_fnError(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory2[*factoryMultiLabel](func(_ context.Context, _, _ string) (*factoryMultiLabel, error) {
		return nil, errTest
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b")
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory2_instanceInitFailure(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory2[TestServiceInterface](func(ctx context.Context, _, _ string) (*TestServiceStruct, error) {
		s := NewMockTestServiceStruct(t)
		s.MockIniter.EXPECT().Init(ctx).Return(errTest)
		return s, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b")
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory2_firstArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory2[*factoryMultiLabel](func(_ context.Context, _ int, _ string) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "not-int", "ok")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory2_FactoryClosures(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	s := pal.ProvideFactory2[*factoryMultiLabel](func(_ context.Context, a string, b int) (*factoryMultiLabel, error) {
		if a == "err" {
			return nil, errTest
		}
		return &factoryMultiLabel{S0: a, I0: b}, nil
	})
	p := newPal(s)
	ctxW := pal.WithPal(ctx, p)
	require.NoError(t, p.Init(ctx))

	f := s.Factory().(func(context.Context, string, int) (*factoryMultiLabel, error))
	got, err := f(ctxW, "hi", 3)
	require.NoError(t, err)
	assert.Equal(t, "hi", got.S0)
	assert.Equal(t, 3, got.I0)

	_, err = f(ctxW, "err", 0)
	assert.ErrorIs(t, err, errTest)

	must := s.MustFactory().(func(context.Context, string, int) *factoryMultiLabel)
	assert.PanicsWithValue(t, errTest, func() { must(ctxW, "err", 0) })
}

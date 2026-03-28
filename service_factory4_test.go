package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

func TestServiceFactory4_Invocation(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory4[*factoryMultiLabel](func(_ context.Context, a string, b int, c string, d bool) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{S0: a, I0: b, S1: c, B: d}, nil
	})
	assert.Equal(t, 4, service.Arguments())

	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	inst, err := p.Invoke(ctx, service.Name(), "a", 1, "c", true)
	require.NoError(t, err)
	l := inst.(*factoryMultiLabel)
	assert.Equal(t, "a", l.S0)
	assert.Equal(t, 1, l.I0)
	assert.Equal(t, "c", l.S1)
	assert.True(t, l.B)

	_, err = p.Invoke(ctx, service.Name(), "a", 1, "c")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)

	_, err = p.Invoke(ctx, service.Name(), "a", 1, "c", "nope")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory4_fnError(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory4[*factoryMultiLabel](func(_ context.Context, _, _, _ string, _ int) (*factoryMultiLabel, error) {
		return nil, errTest
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", "c", 1)
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory4_instanceInitFailure(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory4[TestServiceInterface](func(ctx context.Context, _, _, _ string, _ int) (*TestServiceStruct, error) {
		s := NewMockTestServiceStruct(t)
		s.MockIniter.EXPECT().Init(ctx).Return(errTest)
		return s, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", "c", 1)
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory4_firstArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory4[*factoryMultiLabel](func(_ context.Context, _ int, _, _ string, _ int) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "not-int", "b", "c", 1)
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory4_secondArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory4[*factoryMultiLabel](func(_ context.Context, _ string, _ string, _ string, _ bool) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", 2, "c", true)
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory4_thirdArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory4[*factoryMultiLabel](func(_ context.Context, _, _ string, _ int, _ bool) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", "not-int", true)
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory4_fourthArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory4[*factoryMultiLabel](func(_ context.Context, _, _ string, _ int, _ bool) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", 1, "not-bool")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory4_FactoryClosures(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	s := pal.ProvideFactory4[*factoryMultiLabel](func(_ context.Context, a string, b int, c string, d bool) (*factoryMultiLabel, error) {
		if !d {
			return nil, errTest
		}
		return &factoryMultiLabel{S0: a, I0: b, S1: c, B: d}, nil
	})
	p := newPal(s)
	ctxW := pal.WithPal(ctx, p)
	require.NoError(t, p.Init(ctx))

	f := s.Factory().(func(context.Context, string, int, string, bool) (*factoryMultiLabel, error))
	got, err := f(ctxW, "a", 2, "c", true)
	require.NoError(t, err)
	assert.True(t, got.B)

	_, err = f(ctxW, "a", 2, "c", false)
	assert.ErrorIs(t, err, errTest)

	must := s.MustFactory().(func(context.Context, string, int, string, bool) *factoryMultiLabel)
	assert.PanicsWithValue(t, errTest, func() { must(ctxW, "a", 2, "c", false) })
}

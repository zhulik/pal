package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

func TestServiceFactory5_Invocation(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, a string, b int, c string, d bool, e rune) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{S0: a, I0: b, S1: c, B: d, R: e}, nil
	})
	assert.Equal(t, 5, service.Arguments())

	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	inst, err := p.Invoke(ctx, service.Name(), "a", 2, "c", false, 'z')
	require.NoError(t, err)
	l := inst.(*factoryMultiLabel)
	assert.Equal(t, "a", l.S0)
	assert.Equal(t, 2, l.I0)
	assert.Equal(t, "c", l.S1)
	assert.False(t, l.B)
	assert.Equal(t, 'z', l.R)

	_, err = p.Invoke(ctx, service.Name(), "a", 2, "c", false)
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)

	_, err = p.Invoke(ctx, service.Name(), "a", 2, "c", false, "z")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory5_fnError(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, _ string, _ int, _ string, _ bool, _ rune) (*factoryMultiLabel, error) {
		return nil, errTest
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", 1, "b", true, 'z')
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory5_instanceInitFailure(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[TestServiceInterface](func(ctx context.Context, _ string, _ int, _ string, _ bool, _ rune) (*TestServiceStruct, error) {
		s := NewMockTestServiceStruct(t)
		s.MockIniter.EXPECT().Init(ctx).Return(errTest)
		return s, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", 1, "b", true, 'z')
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory5_firstArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, _ int, _ int, _ string, _ bool, _ rune) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "not-int", 1, "b", true, 'z')
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory5_secondArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, _ string, _ string, _ string, _ bool, _ rune) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", 2, "c", true, 'z')
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory5_thirdArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, _ string, _ int, _ int, _ bool, _ rune) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", 1, "not-int", true, 'z')
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory5_fourthArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, _, _ string, _ int, _ bool, _ rune) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", 1, "not-bool", 'z')
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory5_fifthArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, _, _ string, _ int, _ bool, _ rune) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", 1, true, "not-rune")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory5_FactoryClosures(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	s := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, a string, b int, c string, d bool, e rune) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{S0: a, I0: b, S1: c, B: d, R: e}, nil
	})
	p := newPal(s)
	ctxW := pal.WithPal(ctx, p)
	require.NoError(t, p.Init(ctx))

	f := s.Factory().(func(context.Context, string, int, string, bool, rune) (*factoryMultiLabel, error))
	got, err := f(ctxW, "a", 1, "b", true, 'x')
	require.NoError(t, err)
	assert.Equal(t, 'x', got.R)

	sErr := pal.ProvideFactory5[*factoryMultiLabel](func(_ context.Context, _ string, _ int, _ string, _ bool, _ rune) (*factoryMultiLabel, error) {
		return nil, errTest
	})
	p2 := newPal(sErr)
	ctxW2 := pal.WithPal(ctx, p2)
	require.NoError(t, p2.Init(ctx))

	f2 := sErr.Factory().(func(context.Context, string, int, string, bool, rune) (*factoryMultiLabel, error))
	_, err = f2(ctxW2, "a", 1, "b", true, 'x')
	assert.ErrorIs(t, err, errTest)

	must := sErr.MustFactory().(func(context.Context, string, int, string, bool, rune) *factoryMultiLabel)
	assert.PanicsWithValue(t, errTest, func() { must(ctxW2, "a", 1, "b", true, 'x') })
}

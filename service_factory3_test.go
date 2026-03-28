package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

func TestServiceFactory3_Invocation(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory3[*factoryMultiLabel](func(_ context.Context, a string, b int, c string) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{S0: a, I0: b, S1: c}, nil
	})
	assert.Equal(t, 3, service.Arguments())

	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	inst, err := p.Invoke(ctx, service.Name(), "a", 2, "c")
	require.NoError(t, err)
	l := inst.(*factoryMultiLabel)
	assert.Equal(t, "a", l.S0)
	assert.Equal(t, 2, l.I0)
	assert.Equal(t, "c", l.S1)

	_, err = p.Invoke(ctx, service.Name(), "a", 2)
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentsCount)

	_, err = p.Invoke(ctx, service.Name(), "a", 2, 3)
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory3_fnError(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory3[*factoryMultiLabel](func(_ context.Context, _, _, _ string) (*factoryMultiLabel, error) {
		return nil, errTest
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", "c")
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory3_instanceInitFailure(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory3[TestServiceInterface](func(ctx context.Context, _, _, _ string) (*TestServiceStruct, error) {
		s := NewMockTestServiceStruct(t)
		s.MockIniter.EXPECT().Init(ctx).Return(errTest)
		return s, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", "c")
	assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	assert.ErrorIs(t, err, errTest)
}

func TestServiceFactory3_firstArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory3[*factoryMultiLabel](func(_ context.Context, _ int, _, _ string) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "not-int", "b", "c")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory3_secondArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory3[*factoryMultiLabel](func(_ context.Context, _ string, _ int, _ string) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "not-int", "c")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory3_thirdArgumentWrongType(t *testing.T) {
	t.Parallel()

	service := pal.ProvideFactory3[*factoryMultiLabel](func(_ context.Context, _, _ string, _ int) (*factoryMultiLabel, error) {
		return &factoryMultiLabel{}, nil
	})
	p := newPal(service)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(t.Context()))

	_, err := p.Invoke(ctx, service.Name(), "a", "b", "not-int")
	assert.ErrorIs(t, err, pal.ErrServiceInvalidArgumentType)
}

func TestServiceFactory3_FactoryClosures(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	s := pal.ProvideFactory3[*factoryMultiLabel](func(_ context.Context, a string, b int, c string) (*factoryMultiLabel, error) {
		if a == "err" {
			return nil, errTest
		}
		return &factoryMultiLabel{S0: a, I0: b, S1: c}, nil
	})
	p := newPal(s)
	ctxW := pal.WithPal(ctx, p)
	require.NoError(t, p.Init(ctx))

	f := s.Factory().(func(context.Context, string, int, string) (*factoryMultiLabel, error))
	got, err := f(ctxW, "ok", 1, "c")
	require.NoError(t, err)
	assert.Equal(t, "c", got.S1)

	_, err = f(ctxW, "err", 0, "")
	assert.ErrorIs(t, err, errTest)

	must := s.MustFactory().(func(context.Context, string, int, string) *factoryMultiLabel)
	assert.PanicsWithValue(t, errTest, func() { must(ctxW, "err", 0, "") })
}

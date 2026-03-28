package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

// --- small types for Pal-prefixed lifecycle branches (ServiceFnSingleton.Init / hooks) ---

type fnSingletonPalIniter struct {
	called bool
	err    error
}

func (p *fnSingletonPalIniter) PalInit(ctx context.Context) error {
	p.called = true
	return p.err
}

type fnSingletonIniterOnly struct {
	called bool
	err    error
}

func (i *fnSingletonIniterOnly) Init(ctx context.Context) error {
	i.called = true
	return i.err
}

type fnSingletonPlain struct{ N int }

type fnSingletonRunner struct {
	*MockRunner
}

type fnSingletonPalRunner struct{}

func (*fnSingletonPalRunner) PalRun(ctx context.Context) error {
	return nil
}

type fnSingletonRunConfiger struct {
	cfg *pal.RunConfig
}

func (s *fnSingletonRunConfiger) RunConfig() *pal.RunConfig {
	return s.cfg
}

type fnSingletonPalRunConfiger struct {
	cfg *pal.RunConfig
}

func (s *fnSingletonPalRunConfiger) PalRunConfig() *pal.RunConfig {
	return s.cfg
}

// Implements Runner via Make() template; instance may be nil before Init.
type fnSingletonRunnerFromMake struct{}

func (*fnSingletonRunnerFromMake) Run(ctx context.Context) error {
	return nil
}

type fnSingletonPalRunnerFromMake struct{}

func (*fnSingletonPalRunnerFromMake) PalRun(ctx context.Context) error {
	return nil
}

type fnSingletonPalHealth struct{}

func (*fnSingletonPalHealth) PalHealthCheck(ctx context.Context) error {
	return errTest
}

type fnSingletonPalShutdown struct{}

func (*fnSingletonPalShutdown) PalShutdown(ctx context.Context) error {
	return errTest2
}

// Embeds fnSingletonPalIniter and also implements Initer; PalInit must win in ServiceFnSingleton.Init.
type fnSingletonBothInit struct {
	fnSingletonPalIniter
	initerCalled bool
}

func (b *fnSingletonBothInit) Init(ctx context.Context) error {
	b.initerCalled = true
	return nil
}

func TestServiceFnSingleton_Init(t *testing.T) {
	t.Parallel()

	t.Run("fn returns error", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
			return nil, errTest
		})
		p := newPal(s)
		err := p.Init(pal.WithPal(t.Context(), p))
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("PalInit returns error", func(t *testing.T) {
		t.Parallel()

		inst := &fnSingletonPalIniter{err: errTest}
		s := pal.ProvideFn[*fnSingletonPalIniter](func(_ context.Context) (*fnSingletonPalIniter, error) {
			return inst, nil
		})
		p := newPal(s)
		err := p.Init(pal.WithPal(t.Context(), p))
		assert.ErrorIs(t, err, errTest)
		assert.True(t, inst.called)
	})

	t.Run("Init returns error when only Initer is implemented", func(t *testing.T) {
		t.Parallel()

		inst := &fnSingletonIniterOnly{err: errTest}
		s := pal.ProvideFn[*fnSingletonIniterOnly](func(_ context.Context) (*fnSingletonIniterOnly, error) {
			return inst, nil
		})
		p := newPal(s)
		err := p.Init(pal.WithPal(t.Context(), p))
		assert.ErrorIs(t, err, errTest)
		assert.True(t, inst.called)
	})

	t.Run("PalInit used instead of Init when both would apply", func(t *testing.T) {
		t.Parallel()

		inst := &fnSingletonBothInit{}
		s := pal.ProvideFn[*fnSingletonBothInit](func(_ context.Context) (*fnSingletonBothInit, error) {
			return inst, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(t.Context(), p)))
		assert.True(t, inst.fnSingletonPalIniter.called)
		assert.False(t, inst.initerCalled)
	})

	t.Run("success with PalInit only", func(t *testing.T) {
		t.Parallel()

		inst := &fnSingletonPalIniter{}
		s := pal.ProvideFn[*fnSingletonPalIniter](func(_ context.Context) (*fnSingletonPalIniter, error) {
			return inst, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(t.Context(), p)))
		assert.True(t, inst.called)
	})

	t.Run("success with Init only", func(t *testing.T) {
		t.Parallel()

		inst := &fnSingletonIniterOnly{}
		s := pal.ProvideFn[*fnSingletonIniterOnly](func(_ context.Context) (*fnSingletonIniterOnly, error) {
			return inst, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(t.Context(), p)))
		assert.True(t, inst.called)
	})

	t.Run("success with no init interfaces", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
			return &fnSingletonPlain{N: 42}, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(t.Context(), p)))
	})
}

func TestServiceFnSingleton_Instance(t *testing.T) {
	t.Parallel()

	inst := &fnSingletonPlain{N: 7}
	s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
		return inst, nil
	})
	p := newPal(s)
	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.Init(ctx))

	got, err := s.Instance(ctx)
	require.NoError(t, err)
	assert.Same(t, inst, got)

	viaPal, err := p.Invoke(ctx, s.Name())
	require.NoError(t, err)
	assert.Same(t, inst, viaPal)
}

func TestServiceFnSingleton_Run(t *testing.T) {
	t.Parallel()

	t.Run("no runner", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
			return &fnSingletonPlain{}, nil
		})
		p := newPal(s)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(ctx))
		assert.NoError(t, s.Run(ctx))
	})

	t.Run("Runner", func(t *testing.T) {
		t.Parallel()

		mr := NewMockRunner(t)
		mr.EXPECT().Run(mock.Anything).Return(nil)

		s := pal.ProvideFn[*fnSingletonRunner](func(_ context.Context) (*fnSingletonRunner, error) {
			return &fnSingletonRunner{MockRunner: mr}, nil
		})
		p := newPal(s)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(ctx))
		assert.NoError(t, s.Run(ctx))
	})

	t.Run("PalRunner", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPalRunner](func(_ context.Context) (*fnSingletonPalRunner, error) {
			return &fnSingletonPalRunner{}, nil
		})
		p := newPal(s)
		ctx := pal.WithPal(t.Context(), p)
		require.NoError(t, p.Init(ctx))
		assert.NoError(t, s.Run(ctx))
	})
}

func TestServiceFnSingleton_RunConfig(t *testing.T) {
	t.Parallel()

	custom := &pal.RunConfig{Wait: false}

	t.Run("from instance RunConfiger after Init", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonRunConfiger](func(_ context.Context) (*fnSingletonRunConfiger, error) {
			return &fnSingletonRunConfiger{cfg: custom}, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(t.Context(), p)))
		assert.Equal(t, custom, s.RunConfig())
	})

	t.Run("from instance PalRunConfiger after Init", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPalRunConfiger](func(_ context.Context) (*fnSingletonPalRunConfiger, error) {
			return &fnSingletonPalRunConfiger{cfg: custom}, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(t.Context(), p)))
		assert.Equal(t, custom, s.RunConfig())
	})

	t.Run("default from Make when instance has no config but Make is Runner", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonRunnerFromMake](func(_ context.Context) (*fnSingletonRunnerFromMake, error) {
			return nil, nil
		})
		_ = newPal(s)
		// before Init, instance is nil *fnSingletonRunnerFromMake; Make still produces a concrete *T
		cfg := s.RunConfig()
		require.NotNil(t, cfg)
		assert.True(t, cfg.Wait)
	})

	t.Run("default from Make when Make is PalRunner", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPalRunnerFromMake](func(_ context.Context) (*fnSingletonPalRunnerFromMake, error) {
			return nil, nil
		})
		_ = newPal(s)
		cfg := s.RunConfig()
		require.NotNil(t, cfg)
		assert.True(t, cfg.Wait)
	})

	t.Run("nil when no runner template and plain instance after Init", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
			return &fnSingletonPlain{}, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(t.Context(), p)))
		assert.Nil(t, s.RunConfig())
	})
}

func TestServiceFnSingleton_HealthCheck(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("ToHealthCheck hook runs and overrides PalHealthChecker", func(t *testing.T) {
		t.Parallel()

		inst := &fnSingletonPalHealth{}
		hookCalled := false
		s := pal.ProvideFn[*fnSingletonPalHealth](func(_ context.Context) (*fnSingletonPalHealth, error) {
			return inst, nil
		}).ToHealthCheck(func(_ context.Context, got *fnSingletonPalHealth, _ *pal.Pal) error {
			hookCalled = true
			assert.Same(t, inst, got)
			return errTest
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(ctx, p)))

		err := s.HealthCheck(pal.WithPal(ctx, p))
		assert.ErrorIs(t, err, errTest)
		assert.True(t, hookCalled)
	})

	t.Run("PalHealthCheck when no hook", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPalHealth](func(_ context.Context) (*fnSingletonPalHealth, error) {
			return &fnSingletonPalHealth{}, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(ctx, p)))
		assert.ErrorIs(t, s.HealthCheck(pal.WithPal(ctx, p)), errTest)
	})

	t.Run("HealthChecker when no hook", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*TestServiceStruct](func(_ context.Context) (*TestServiceStruct, error) {
			out := NewMockTestServiceStruct(t)
			out.MockIniter.EXPECT().Init(mock.Anything).Return(nil)
			out.MockHealthChecker.EXPECT().HealthCheck(mock.Anything).Return(errTest)
			return out, nil
		})
		p := newPal(s)
		wrapped := pal.WithPal(ctx, p)
		require.NoError(t, p.Init(wrapped))
		assert.ErrorIs(t, s.HealthCheck(wrapped), errTest)
	})

	t.Run("noop when no hook and no checker", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
			return &fnSingletonPlain{}, nil
		})
		p := newPal(s)
		wrapped := pal.WithPal(ctx, p)
		require.NoError(t, p.Init(wrapped))
		assert.NoError(t, s.HealthCheck(wrapped))
	})
}

func TestServiceFnSingleton_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("ToShutdown hook runs and overrides PalShutdowner", func(t *testing.T) {
		t.Parallel()

		inst := &fnSingletonPalShutdown{}
		hookCalled := false
		s := pal.ProvideFn[*fnSingletonPalShutdown](func(_ context.Context) (*fnSingletonPalShutdown, error) {
			return inst, nil
		}).ToShutdown(func(_ context.Context, got *fnSingletonPalShutdown, _ *pal.Pal) error {
			hookCalled = true
			assert.Same(t, inst, got)
			return errTest
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(ctx, p)))

		err := s.Shutdown(pal.WithPal(ctx, p))
		assert.ErrorIs(t, err, errTest)
		assert.True(t, hookCalled)
	})

	t.Run("PalShutdown when no hook", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPalShutdown](func(_ context.Context) (*fnSingletonPalShutdown, error) {
			return &fnSingletonPalShutdown{}, nil
		})
		p := newPal(s)
		require.NoError(t, p.Init(pal.WithPal(ctx, p)))
		assert.ErrorIs(t, s.Shutdown(pal.WithPal(ctx, p)), errTest2)
	})

	t.Run("Shutdown when no hook", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*TestServiceStruct](func(_ context.Context) (*TestServiceStruct, error) {
			out := NewMockTestServiceStruct(t)
			out.MockIniter.EXPECT().Init(mock.Anything).Return(nil)
			out.MockShutdowner.EXPECT().Shutdown(mock.Anything).Return(errTest)
			return out, nil
		})
		p := newPal(s)
		wrapped := pal.WithPal(ctx, p)
		require.NoError(t, p.Init(wrapped))
		assert.ErrorIs(t, s.Shutdown(wrapped), errTest)
	})

	t.Run("noop when no hook and no shutdowner", func(t *testing.T) {
		t.Parallel()

		s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
			return &fnSingletonPlain{}, nil
		})
		p := newPal(s)
		wrapped := pal.WithPal(ctx, p)
		require.NoError(t, p.Init(wrapped))
		assert.NoError(t, s.Shutdown(wrapped))
	})
}

func TestServiceFnSingleton_ToShutdown_ToHealthCheck_returnSelf(t *testing.T) {
	t.Parallel()

	s := pal.ProvideFn[*fnSingletonPlain](func(_ context.Context) (*fnSingletonPlain, error) {
		return &fnSingletonPlain{}, nil
	})
	s2 := s.ToShutdown(func(_ context.Context, _ *fnSingletonPlain, _ *pal.Pal) error {
		return nil
	})
	assert.Same(t, s, s2)

	s3 := s.ToHealthCheck(func(_ context.Context, _ *fnSingletonPlain, _ *pal.Pal) error {
		return nil
	})
	assert.Same(t, s, s3)
}

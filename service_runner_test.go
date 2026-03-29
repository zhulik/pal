package pal_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

func TestServiceRunner_RunConfig(t *testing.T) {
	t.Parallel()

	r := pal.ProvideRunner(func(context.Context) error { return nil })
	cfg := r.RunConfig()

	require.NotNil(t, cfg)
	assert.True(t, cfg.Wait)
}

func TestServiceRunner_Run(t *testing.T) {
	t.Parallel()

	t.Run("invokes fn with context", func(t *testing.T) {
		t.Parallel()

		var got context.Context
		r := pal.ProvideRunner(func(ctx context.Context) error {
			got = ctx
			return nil
		})
		r.P = newPal(r)
		ctx := t.Context()
		require.NoError(t, r.Run(ctx))
		assert.Equal(t, ctx, got)
	})

	t.Run("returns fn error", func(t *testing.T) {
		t.Parallel()

		r := pal.ProvideRunner(func(context.Context) error { return errTest })
		r.P = newPal(r)
		assert.ErrorIs(t, r.Run(t.Context()), errTest)
	})

	t.Run("returns PanicError when fn panics", func(t *testing.T) {
		t.Parallel()

		const msg = "runner panic"
		r := pal.ProvideRunner(func(context.Context) error { panic(msg) })
		r.P = newPal(r)
		err := r.Run(t.Context())
		require.Error(t, err)
		var panicErr *pal.PanicError
		require.True(t, errors.As(err, &panicErr))
		assert.Contains(t, panicErr.Error(), msg)
		assert.NotEmpty(t, panicErr.Backtrace())
	})
}

func TestServiceRunner_Instance(t *testing.T) {
	t.Parallel()

	r := pal.ProvideRunner(func(context.Context) error { return nil })
	inst, err := r.Instance(t.Context())

	assert.NoError(t, err)
	assert.Nil(t, inst)
}

func TestServiceRunner_Name(t *testing.T) {
	t.Parallel()

	const prefix = "$function-runner-"
	r := pal.ProvideRunner(func(context.Context) error { return nil })
	name := r.Name()

	assert.True(t, strings.HasPrefix(name, prefix))
	suffix := strings.TrimPrefix(name, prefix)
	assert.Len(t, suffix, 8)
	for _, c := range suffix {
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
		assert.True(t, ok, "unexpected char %q in suffix %q", c, suffix)
	}
}

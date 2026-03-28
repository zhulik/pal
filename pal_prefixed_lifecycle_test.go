package pal_test

import (
	"context"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

// palOnlyLifecycle implements only Pal-prefixed lifecycle methods.
type palOnlyLifecycle struct {
	palInit, palHealth, palShutdown bool
}

func (s *palOnlyLifecycle) PalInit(_ context.Context) error {
	s.palInit = true
	return nil
}

func (s *palOnlyLifecycle) PalHealthCheck(_ context.Context) error {
	s.palHealth = true
	return nil
}

func (s *palOnlyLifecycle) PalShutdown(_ context.Context) error {
	s.palShutdown = true
	return nil
}

func TestPalPrefixed_palOnlyInitHealthShutdown(t *testing.T) {
	t.Parallel()

	s := &palOnlyLifecycle{}
	p := newPal(pal.Provide(s))

	require.NoError(t, p.Init(t.Context()))
	assert.True(t, s.palInit, "PalInit should run when only PalIniter is implemented")

	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.HealthCheck(ctx))
	assert.True(t, s.palHealth, "PalHealthCheck should run when only PalHealthChecker is implemented")

	require.NoError(t, p.Container().Shutdown(ctx))
	assert.True(t, s.palShutdown, "PalShutdown should run when only PalShutdowner is implemented")
}

// dualInitHealthShutdown implements both standard and Pal-prefixed lifecycle methods.
type dualInitHealthShutdown struct {
	palInit, stdInit           bool
	palHealth, stdHealth       bool
	palShutdown, stdShutdown   bool
	palRunCalled, stdRunCalled bool
}

func (d *dualInitHealthShutdown) PalInit(_ context.Context) error {
	d.palInit = true
	return nil
}

func (d *dualInitHealthShutdown) Init(_ context.Context) error {
	d.stdInit = true
	return nil
}

func (d *dualInitHealthShutdown) PalHealthCheck(_ context.Context) error {
	d.palHealth = true
	return nil
}

func (d *dualInitHealthShutdown) HealthCheck(_ context.Context) error {
	d.stdHealth = true
	return nil
}

func (d *dualInitHealthShutdown) PalShutdown(_ context.Context) error {
	d.palShutdown = true
	return nil
}

func (d *dualInitHealthShutdown) Shutdown(_ context.Context) error {
	d.stdShutdown = true
	return nil
}

func (d *dualInitHealthShutdown) PalRun(_ context.Context) error {
	d.palRunCalled = true
	return nil
}

func (d *dualInitHealthShutdown) Run(_ context.Context) error {
	d.stdRunCalled = true
	return nil
}

func TestPalPrefixed_precedenceOverStandardMethods(t *testing.T) {
	t.Parallel()

	s := &dualInitHealthShutdown{}
	p := newPal(pal.Provide(s))

	require.NoError(t, p.Init(t.Context()))
	assert.True(t, s.palInit)
	assert.False(t, s.stdInit)

	ctx := pal.WithPal(t.Context(), p)
	require.NoError(t, p.HealthCheck(ctx))
	assert.True(t, s.palHealth)
	assert.False(t, s.stdHealth)

	require.NoError(t, p.Container().Shutdown(ctx))
	assert.True(t, s.palShutdown)
	assert.False(t, s.stdShutdown)

	require.NoError(t, p.Run(t.Context(), syscall.SIGINT))
	assert.True(t, s.palRunCalled)
	assert.False(t, s.stdRunCalled)
}

type palInitWithHook struct {
	palInitCalled bool
}

func (s *palInitWithHook) PalInit(_ context.Context) error {
	s.palInitCalled = true
	return nil
}

func TestPalPrefixed_ToInitHookOverridesPalInit(t *testing.T) {
	t.Parallel()

	s := &palInitWithHook{}
	hookCalled := false

	p := newPal(pal.Provide(s).ToInit(func(_ context.Context, _ *palInitWithHook, _ *pal.Pal) error {
		hookCalled = true
		return nil
	}))

	require.NoError(t, p.Init(t.Context()))
	assert.True(t, hookCalled)
	assert.False(t, s.palInitCalled, "PalInit must not run when ToInit hook is set")
}

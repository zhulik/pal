package pal_test

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zhulik/pal"
)

type MainRunner any

func TestRunServices(t *testing.T) {
	t.Run("returns error if no main runners are given", func(t *testing.T) {
		t.Parallel()

		secondaryRunner := pal.ProvideFn[*RunnerServiceStruct](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: false})
			return s, nil
		})

		p := newPal(secondaryRunner)
		assert.NoError(t, p.Init(t.Context()))

		err := pal.RunServices(t.Context(), slices.Collect(maps.Values(p.Services())))
		assert.ErrorIs(t, err, pal.ErrNoMainRunners)
	})

	t.Run("returns nil if main and secondary runners finish successfully", func(t *testing.T) {
		t.Parallel()

		mainRunner := pal.ProvideFn[MainRunner](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: true})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(nil)
			return s, nil
		})

		secondaryRunner := pal.ProvideFn[*RunnerServiceStruct](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: false})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(nil)
			return s, nil
		})

		p := newPal(mainRunner, secondaryRunner)
		assert.NoError(t, p.Init(t.Context()))

		err := pal.RunServices(t.Context(), slices.Collect(maps.Values(p.Services())))
		assert.NoError(t, err)
	})

	t.Run("returns nil if multiple main runners finish successfully", func(t *testing.T) {
		t.Parallel()

		mainRunner1 := pal.ProvideFn[MainRunner](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: true})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(nil)
			return s, nil
		})

		mainRunner2 := pal.ProvideFn[MainRunner](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: true})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(nil)
			return s, nil
		})

		p := newPal(mainRunner1, mainRunner2)
		assert.NoError(t, p.Init(t.Context()))

		err := pal.RunServices(t.Context(), slices.Collect(maps.Values(p.Services())))
		assert.NoError(t, err)
	})

	t.Run("returns err if main runners blocks and secondary runner fails", func(t *testing.T) {
		t.Parallel()

		var mainCompleted bool

		mainRunner := pal.ProvideFn[MainRunner](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: true})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(context.Canceled).Run(func(ctx context.Context) {
				<-ctx.Done()
				mainCompleted = true
			})
			return s, nil
		})

		secondaryRunner := pal.ProvideFn[*RunnerServiceStruct](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: false})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(errTest)
			return s, nil
		})

		p := newPal(mainRunner, secondaryRunner)
		assert.NoError(t, p.Init(t.Context()))

		err := pal.RunServices(t.Context(), slices.Collect(maps.Values(p.Services())))
		assert.ErrorIs(t, err, errTest)
		assert.True(t, mainCompleted)
	})

	t.Run("returns err if main runner finishes successfully and secondary runner blocks", func(t *testing.T) {
		t.Parallel()

		var secondaryCompleted bool

		mainRunner := pal.ProvideFn[MainRunner](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: true})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(nil)
			return s, nil
		})

		secondaryRunner := pal.ProvideFn[*RunnerServiceStruct](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: false})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(context.Canceled).Run(func(ctx context.Context) {
				<-ctx.Done()
				secondaryCompleted = true
			})
			return s, nil
		})

		p := newPal(mainRunner, secondaryRunner)
		assert.NoError(t, p.Init(t.Context()))

		err := pal.RunServices(t.Context(), slices.Collect(maps.Values(p.Services())))
		assert.NoError(t, err)
		assert.True(t, secondaryCompleted)
	})

	t.Run("returns err if main runner fails and secondary runner blocks", func(t *testing.T) {
		t.Parallel()

		var secondaryCompleted bool

		mainRunner := pal.ProvideFn[MainRunner](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: true})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(errTest)
			return s, nil
		})

		secondaryRunner := pal.ProvideFn[*RunnerServiceStruct](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: false})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(context.Canceled).Run(func(ctx context.Context) {
				<-ctx.Done()
				secondaryCompleted = true
			})
			return s, nil
		})

		p := newPal(mainRunner, secondaryRunner)
		assert.NoError(t, p.Init(t.Context()))

		err := pal.RunServices(t.Context(), slices.Collect(maps.Values(p.Services())))
		assert.ErrorIs(t, err, errTest)
		assert.True(t, secondaryCompleted)
	})

	t.Run("returns a joined err if main main and secondary runners fail", func(t *testing.T) {
		t.Parallel()

		mainRunner := pal.ProvideFn[MainRunner](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: true})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(errTest)
			return s, nil
		})

		secondaryRunner := pal.ProvideFn[*RunnerServiceStruct](func(context.Context) (*RunnerServiceStruct, error) {
			s := NewMockRunnerServiceStruct(t)
			s.MockRunConfiger.EXPECT().RunConfig().Return(&pal.RunConfig{Wait: false})
			s.MockRunner.EXPECT().Run(mock.Anything).Return(errTest2)
			return s, nil
		})

		p := newPal(mainRunner, secondaryRunner)
		assert.NoError(t, p.Init(t.Context()))

		err := pal.RunServices(t.Context(), slices.Collect(maps.Values(p.Services())))
		assert.ErrorIs(t, err, errTest)
		assert.ErrorIs(t, err, errTest2)
	})
}

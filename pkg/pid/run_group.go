package pid

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/alphadose/haxmap"
	"golang.org/x/sync/errgroup"
)

type jobFN func(context.Context) error

type RunGroup struct {
	mainWG      *errgroup.Group
	secondaryWG *errgroup.Group

	counter atomic.Int32
	jobs    *haxmap.Map[int32, *job]

	wait func() error
	stop func()
}

func NewRunGroup() *RunGroup {
	mainWG, mainCtx := errgroup.WithContext(context.Background())
	secondaryWG, secondaryCtx := errgroup.WithContext(context.Background())

	var g *RunGroup
	g = &RunGroup{
		mainWG:      mainWG,
		secondaryWG: secondaryWG,
		jobs:        haxmap.New[int32, *job](),
		wait:        sync.OnceValue(func() error { return errors.Join(mainWG.Wait(), secondaryWG.Wait()) }),
		stop: sync.OnceFunc(func() {
			g.jobs.ForEach(func(_ int32, j *job) bool {
				j.cancel()
				return true
			})
		}),
	}

	go func() {
		select {
		case <-mainCtx.Done():
		case <-secondaryCtx.Done():
		}
		g.stop()
	}()

	return g
}

func (g *RunGroup) Go(ctx context.Context, wait bool, fn jobFN) {
	fnCtx, cancel := context.WithCancel(ctx)

	j := &job{
		fn:     fn,
		wait:   wait,
		cancel: cancel,
	}

	g.jobs.Set(g.counter.Add(1), j)

	var wg *errgroup.Group

	if wait {
		wg = g.mainWG
	} else {
		wg = g.secondaryWG
	}

	wg.Go(func() error {
		err := j.run(fnCtx)

		if err != nil {
			return err
		}

		running := 0
		g.jobs.ForEach(func(_ int32, j *job) bool {
			if !j.wait {
				return true
			}
			running++
			if j.done.Load() {
				running--
			}
			return true
		})

		if running == 0 {
			g.stop()
		}
		return nil
	})
}

func (g *RunGroup) Wait() error {
	return g.wait()
}

func (g *RunGroup) Stop() {
	g.stop()
}

type job struct {
	fn     jobFN
	wait   bool
	cancel context.CancelFunc

	done atomic.Bool
}

func (j *job) run(ctx context.Context) error {
	defer j.done.Store(true)

	return j.fn(ctx)
}

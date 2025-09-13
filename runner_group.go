package pal

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type RunnerGroup struct {
	main      *errgroup.Group
	secondary *errgroup.Group

	cancel func(error)
}

// Run runs the services in 2 runner groups: main and secondary.
// It returns true if any main runners started.
// If no main runners among the services, it returns false and no error and does nothing.
// if any of the runners fail, the error is returned and and all other runners are stopped
// by cancelling the context.
func (r *RunnerGroup) Run(ctx context.Context, services []ServiceDef) (bool, error) {
	mainRunners := []ServiceDef{}
	secondaryRunners := []ServiceDef{}

	for _, service := range services {
		runCfg := service.RunConfig()

		// run config is nil if the service is not a runner
		if runCfg == nil {
			continue
		}

		if runCfg.Wait {
			mainRunners = append(mainRunners, service)
		} else {
			secondaryRunners = append(secondaryRunners, service)
		}
	}

	if len(mainRunners) == 0 {
		return false, nil
	}

	ctx, r.cancel = context.WithCancelCause(ctx)
	defer r.cancel(nil)

	var mainCtx context.Context
	r.main, mainCtx = errgroup.WithContext(ctx)
	go func() {
		// Errgoup cancels the context when any of the tasks it manages returns an error.
		// We want to stop if any of the runners fails no matter main or secondary and propagate
		// the cause to the main context.
		<-mainCtx.Done()
		// propagate the cause to the main context
		r.cancel(context.Cause(mainCtx))
	}()

	for _, service := range mainRunners {
		r.main.Go(func() error {
			return service.Run(ctx)
		})
	}

	if len(secondaryRunners) > 0 {
		var secondaryCtx context.Context
		r.secondary, secondaryCtx = errgroup.WithContext(ctx)
		go func() {
			<-secondaryCtx.Done()
			r.cancel(context.Cause(secondaryCtx))
		}()
	}

	for _, service := range secondaryRunners {
		r.secondary.Go(func() error {
			return service.Run(ctx)
		})
	}

	<-ctx.Done()

	return true, context.Cause(ctx)
}

func (r *RunnerGroup) Stop(ctx context.Context) error {
	if r.cancel == nil {
		return nil
	}
	return r.main.Wait()
}

package pal

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
)

var ErrNoMainRunners = errors.New("no main runners found")

// RunServices runs the services in 2 runner groups: main and secondary.
// Block until:
// - passed context is canceled
// - any of the runners fails
// - all runners finish their work
// It returns ErrNoMainRunners if no main runners among the services.
// if any of the runners fail, the error is returned and and all other runners are stopped
// by cancelling the context passed to them.
func RunServices(ctx context.Context, services []ServiceDef) error {
	mainRunners, secondaryRunners := getRunners(services)

	if len(mainRunners) == 0 {
		return ErrNoMainRunners
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	runService := func(service ServiceDef) func() error {
		return func() error {
			err := service.Run(ctx)
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
	}

	awaitGroupContextAndCancelRoot := func(groupCtx context.Context) {
		go func() {
			// Errgoup cancels the context when any of the tasks it manages returns an error.
			// We want to stop if any of the runners fails no matter main or secondary and propagate
			// the cause to the main context.
			<-groupCtx.Done()

			cancel()
		}()
	}

	main, mainCtx := errgroup.WithContext(ctx)
	go awaitGroupContextAndCancelRoot(mainCtx)
	for _, service := range mainRunners {
		main.Go(runService(service))
	}

	var secondary *errgroup.Group
	if len(secondaryRunners) > 0 {
		var secondaryCtx context.Context
		secondary, secondaryCtx = errgroup.WithContext(ctx)
		go awaitGroupContextAndCancelRoot(secondaryCtx)
	}

	for _, service := range secondaryRunners {
		secondary.Go(runService(service))
	}

	// block until all main and secondary(if any) runners finish
	err := main.Wait()
	if secondary != nil {
		err = errors.Join(err, secondary.Wait())
	}

	return err
}

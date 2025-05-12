package pal

import (
	"context"
)

func buildService[S any](ctx context.Context, beforeInit LifecycleHook[S], name string) (*S, error) {
	p := FromContext(ctx)

	logger := p.logger.With("service", name)

	logger.Debug("Creating an instance")
	s, err := Inject[S](ctx, FromContext(ctx))
	if err != nil {
		return nil, err
	}

	if beforeInit != nil {
		logger.Debug("Calling BeforeInit hook")
		err = beforeInit(ctx, s)
		if err != nil {
			logger.Warn("BeforeInit returned error", "error", err)
			return nil, err
		}
	}

	if initer, ok := any(s).(Initer); ok {
		logger.Debug("Calling Init method")
		if err := initer.Init(ctx); err != nil {
			logger.Warn("Init returned error", "error", err)
			return nil, err
		}
	}
	return s, nil
}

func runService(ctx context.Context, instance any, name string) error {
	p := FromContext(ctx)
	logger := p.logger.With("service", name)

	runner, ok := instance.(Runner)
	if !ok {
		return nil
	}

	return tryWrap(func() error {
		logger.Info("Running")
		err := runner.Run(ctx)
		if err != nil {
			logger.Warn("Runner exited with error, scheduling shutdown", "error", err)
			FromContext(ctx).Shutdown(err)
			return err
		}

		logger.Info("Runner finished successfully")
		return nil
	})()
}

func healthcheckService(ctx context.Context, instance any) error {
	if h, ok := instance.(HealthChecker); ok {
		return h.HealthCheck(ctx)
	}
	return nil
}

func shutdownService(ctx context.Context, instance any) error {
	if h, ok := instance.(Shutdowner); ok {
		return h.Shutdown(ctx)
	}
	return nil
}

package pal

import (
	"context"
	"log/slog"
)

func buildService[S any](ctx context.Context, beforeInit LifecycleHook[*S], p *Pal, logger *slog.Logger) (*S, error) {
	logger.Debug("Creating an instance")
	s, err := Build[S](ctx, p)
	if err != nil {
		return nil, err
	}

	if beforeInit != nil {
		logger.Info("Calling BeforeInit hook")
		err = beforeInit(ctx, s)
		if err != nil {
			p.logger.Warn("BeforeInit hook failed", "error", err)
			return nil, err
		}
	}

	if initer, ok := any(s).(Initer); ok {
		logger.Info("Calling Init method")
		if err := initer.Init(ctx); err != nil {
			logger.Warn("Init failed", "error", err)
			return nil, err
		}
	}
	return s, nil
}

func runService(ctx context.Context, instance any, logger *slog.Logger) error {
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

func healthcheckService(ctx context.Context, instance any, logger *slog.Logger) error {
	h, ok := instance.(HealthChecker)
	if !ok {
		return nil
	}

	err := h.HealthCheck(ctx)
	if err != nil {
		logger.Warn("Healthcheck failed", "error", err)
		return err
	}

	return nil
}

func shutdownService(ctx context.Context, instance any, logger *slog.Logger) error {
	h, ok := instance.(Shutdowner)
	if !ok {
		return nil
	}

	err := h.Shutdown(ctx)
	if err != nil {
		logger.Warn("Shutdown failed", "error", err)
		return err
	}

	return nil
}

package pal

import (
	"context"
)

func runService(ctx context.Context, name string, instance any, p *Pal) error {
	logger := p.logger.With("service", name)
	runner, ok := instance.(Runner)
	if !ok {
		return nil
	}

	return tryWrap(func() error {
		logger.Debug("Running")
		err := runner.Run(ctx)
		if err != nil {
			logger.Error("Runner exited with error, scheduling shutdown", "error", err)
			FromContext(ctx).Shutdown(err)
			return err
		}

		logger.Debug("Runner finished successfully")
		return nil
	})()
}

func runConfig(instance any) *RunConfig {
	configer, ok := instance.(RunConfiger)
	if ok {
		return configer.RunConfig()
	}

	return nil
}

func healthcheckService[T any](ctx context.Context, name string, instance T, hook LifecycleHook[T], p *Pal) error {
	logger := p.logger.With("service", name)
	if hook != nil {
		logger.Debug("Calling ToHealthCheck hook")
		err := hook(ctx, instance, p)
		if err != nil {
			logger.Error("Healthcheck hook failed", "error", err)
		}
		return err
	}

	h, ok := any(instance).(HealthChecker)
	if !ok {
		return nil
	}

	err := h.HealthCheck(ctx)
	if err != nil {
		logger.Error("Healthcheck failed", "error", err)
		return err
	}

	return nil
}

func shutdownService[T any](ctx context.Context, name string, instance T, hook LifecycleHook[T], p *Pal) error {
	logger := p.logger.With("service", name)
	if hook != nil {
		logger.Debug("Calling ToShutdown hook")
		err := hook(ctx, instance, p)
		if err != nil {
			logger.Error("Shutdown hook failed", "error", err)
		}
		return err
	}

	h, ok := any(instance).(Shutdowner)
	if !ok {
		return nil
	}

	err := h.Shutdown(ctx)
	if err != nil {
		logger.Error("Shutdown failed", "error", err)
		return err
	}

	return nil
}

func initService[T any](ctx context.Context, name string, instance T, hook LifecycleHook[T], p *Pal) error {
	logger := p.logger.With("service", name)

	if hook != nil {
		logger.Debug("Calling ToInit hook")
		err := hook(ctx, instance, p)
		if err != nil {
			logger.Error("Init hook failed", "error", err)
		}
		return err
	}

	if initer, ok := any(instance).(Initer); ok && any(instance) != any(p) {
		logger.Debug("Calling Init method")
		if err := initer.Init(ctx); err != nil {
			logger.Error("Init failed", "error", err)
			return err
		}
	}
	return nil
}

package pal

import (
	"context"
	"log/slog"
)

func runService(ctx context.Context, instance any, logger *slog.Logger) error {
	runner, ok := instance.(Runner)
	if !ok {
		return nil
	}

	return tryWrap(func() error {
		logger.Info("Running")
		err := runner.Run(ctx)
		if err != nil {
			logger.Error("Runner exited with error, scheduling shutdown", "error", err)
			FromContext(ctx).Shutdown(err)
			return err
		}

		logger.Info("Runner finished successfully")
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

func healthcheckService(ctx context.Context, instance any, logger *slog.Logger) error {
	h, ok := instance.(HealthChecker)
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

func shutdownService(ctx context.Context, instance any, logger *slog.Logger) error {
	h, ok := instance.(Shutdowner)
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

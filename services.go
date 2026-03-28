package pal

import (
	"context"
	"fmt"
)

func runService(ctx context.Context, name string, instance any, p *Pal) error {
	logger := p.logger.With("service", name)

	var run func() error
	if pr, ok := instance.(PalRunner); ok {
		run = func() error {
			logger.Debug("Running")
			err := pr.PalRun(ctx)
			if err != nil {
				logger.Error("Runner exited with error", "error", err)
				return err
			}
			logger.Debug("Runner finished successfully")
			return nil
		}
	} else if runner, ok := instance.(Runner); ok {
		run = func() error {
			logger.Debug("Running")
			err := runner.Run(ctx)
			if err != nil {
				logger.Error("Runner exited with error", "error", err)
				return err
			}
			logger.Debug("Runner finished successfully")
			return nil
		}
	} else {
		return nil
	}

	err := tryWrap(run)()

	if err != nil {
		if panicErr, ok := err.(*PanicError); ok {
			fmt.Printf("panic: %s\n%s\n", panicErr.Error(), panicErr.Backtrace())
		}
	}

	return err
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

	if ph, ok := any(instance).(PalHealthChecker); ok {
		err := ph.PalHealthCheck(ctx)
		if err != nil {
			logger.Error("Healthcheck failed", "error", err)
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

	if ps, ok := any(instance).(PalShutdowner); ok {
		err := ps.PalShutdown(ctx)
		if err != nil {
			logger.Error("Shutdown failed", "error", err)
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

	err := p.InjectInto(ctx, instance)
	if err != nil {
		return err
	}

	if hook != nil {
		logger.Debug("Calling ToInit hook")
		err := hook(ctx, instance, p)
		if err != nil {
			logger.Error("Init hook failed", "error", err)
		}
		return err
	}

	if pi, ok := any(instance).(PalIniter); ok && any(instance) != any(p) {
		logger.Debug("Calling PalInit method")
		if err := pi.PalInit(ctx); err != nil {
			logger.Error("Init failed", "error", err)
			return err
		}
		return nil
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

func flattenServices(services []ServiceDef) []ServiceDef {
	seen := make(map[ServiceDef]bool)
	var result []ServiceDef

	var process func([]ServiceDef)
	process = func(svcs []ServiceDef) {
		for _, svc := range svcs {
			if _, ok := seen[svc]; !ok {
				seen[svc] = true

				if _, ok := svc.(*ServiceList); !ok {
					result = append(result, svc)
				}

				process(svc.Dependencies())
			}
		}
	}

	process(services)
	return result
}

// palOrStandardRunConfig returns scheduling config from a service instance when it implements
// [PalRunConfiger] or [RunConfiger]; otherwise nil.
func palOrStandardRunConfig(instance any) *RunConfig {
	if pc, ok := instance.(PalRunConfiger); ok {
		return pc.PalRunConfig()
	}
	if c, ok := instance.(RunConfiger); ok {
		return c.RunConfig()
	}
	return nil
}

func getRunners(services []ServiceDef) ([]ServiceDef, []ServiceDef) {
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

	return mainRunners, secondaryRunners
}

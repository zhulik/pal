package pal

import (
	"context"
)

type ServiceSingleton[I any, S any] struct {
	beforeInit LifecycleHook[S]
	instance   I
}

func (s *ServiceSingleton[I, S]) Run(ctx context.Context) error {
	return runService(ctx, s.instance, s.Name())
}

func (s *ServiceSingleton[I, S]) Name() string {
	return elem[I]().String()
}

// Init creates a new instance of the Service, calls its Init method if it implements Initer.
func (s *ServiceSingleton[I, S]) Init(ctx context.Context) error {
	instance, err := buildInstance[S](ctx, s.beforeInit, s.Name())
	if err != nil {
		return err
	}

	// it is cast here to make sure it explodes during init
	s.instance = any(instance).(I)

	return nil
}

func (s *ServiceSingleton[I, S]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, s.instance)
}

func (s *ServiceSingleton[I, S]) Shutdown(ctx context.Context) error {
	return shutdownService(ctx, s.instance)
}

func shutdownService(ctx context.Context, instance any) error {
	if h, ok := instance.(Shutdowner); ok {
		return h.Shutdown(ctx)
	}
	return nil
}

func (s *ServiceSingleton[I, S]) Make() any {
	return new(S)
}

func (s *ServiceSingleton[I, S]) Instance(_ context.Context) (any, error) {
	return s.instance, nil
}

func (s *ServiceSingleton[I, S]) BeforeInit(hook LifecycleHook[S]) *ServiceSingleton[I, S] {
	s.beforeInit = hook
	return s
}

func (s *ServiceSingleton[I, S]) Validate(ctx context.Context) error {
	return validateService[I, S](ctx)
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

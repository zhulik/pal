package pal

import (
	"context"
)

type Service[I any, S any] struct {
	beforeInit LifecycleHook[S]
	instance   I
}

func (s *Service[I, S]) Run(ctx context.Context) error {
	return runService(ctx, s.instance, s.Name())
}

func (s *Service[I, S]) Name() string {
	return elem[I]().String()
}

// Init creates a new instance of the Service, calls its Init method if it implements Initer.
func (s *Service[I, S]) Init(ctx context.Context) error {
	instance, err := buildInstance[S](ctx, s.beforeInit, s.Name())
	if err != nil {
		return err
	}

	// it is cast here to make sure it explodes during init
	s.instance = any(instance).(I)

	return nil
}

func (s *Service[I, S]) HealthCheck(ctx context.Context) error {
	if h, ok := any(s.instance).(HealthChecker); ok {
		return h.HealthCheck(ctx)
	}
	return nil
}

func (s *Service[I, S]) Shutdown(ctx context.Context) error {
	if h, ok := any(s.instance).(Shutdowner); ok {
		return h.Shutdown(ctx)
	}
	return nil
}

func (s *Service[I, S]) Make() any {
	return new(S)
}

func (s *Service[I, S]) Instance(_ context.Context) (any, error) {
	return s.instance, nil
}

func (s *Service[I, S]) BeforeInit(hook LifecycleHook[S]) *Service[I, S] {
	s.beforeInit = hook
	return s
}

func (s *Service[I, S]) Validate(ctx context.Context) error {
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

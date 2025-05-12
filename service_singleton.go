package pal

import (
	"context"
)

type ServiceSingleton[I any, S any] struct {
	P          *Pal
	beforeInit LifecycleHook[S]
	instance   I
}

func (s *ServiceSingleton[I, S]) Run(ctx context.Context) error {
	return runService(ctx, s.instance, s.P.logger.With("service", s.Name()))
}

func (s *ServiceSingleton[I, S]) Name() string {
	return elem[I]().String()
}

// Init creates a new instance of the Service, calls its Init method if it implements Initer.
func (s *ServiceSingleton[I, S]) Init(ctx context.Context) error {
	instance, err := buildService[S](ctx, s.beforeInit, s.P, s.P.logger.With("service", s.Name()))
	if err != nil {
		return err
	}

	// it is cast here to make sure it explodes during init
	s.instance = any(instance).(I)

	return nil
}

func (s *ServiceSingleton[I, S]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, s.instance, s.P.logger.With("service", s.Name()))
}

func (s *ServiceSingleton[I, S]) Shutdown(ctx context.Context) error {
	return shutdownService(ctx, s.instance, s.P.logger.With("service", s.Name()))
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

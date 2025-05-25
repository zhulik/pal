package pal

import (
	"context"
)

// ServiceSingleton represents a singleton service in the container.
// A singleton service is created only once during application initialization and reused for all subsequent requests.
// I is the interface type that the service will be cast to, and S is the concrete implementation type.
// Typically, I would be an interface that S implements, or a pointer to S.
type ServiceSingleton[I any, S any] struct {
	P          *Pal
	beforeInit LifecycleHook[*S]
	instance   I
}

func (s *ServiceSingleton[I, S]) Dependencies() []ServiceDef {
	return nil
}

// Run implements the Runner interface.
// It delegates to the underlying service instance if it implements the Runner interface.
func (s *ServiceSingleton[I, S]) Run(ctx context.Context) error {
	return runService(ctx, s.instance, s.P.logger.With("service", s.Name()))
}

// Name returns the name of the service, which is derived from the interface type I.
func (s *ServiceSingleton[I, S]) Name() string {
	return elem[I]().String()
}

// Init implements the Initer interface.
// It creates a new instance of the service, injects its dependencies, and calls its Init method if it implements Initer.
// The instance is then cast to type I and stored for future use.
func (s *ServiceSingleton[I, S]) Init(ctx context.Context) error {
	instance, err := buildService[S](ctx, s.beforeInit, s.P, s.P.logger.With("service", s.Name()))
	if err != nil {
		return err
	}

	// it is cast here to make sure it explodes during init
	s.instance = any(instance).(I)

	return nil
}

// HealthCheck implements the HealthChecker interface.
// It delegates to the underlying service instance if it implements the HealthChecker interface.
func (s *ServiceSingleton[I, S]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, s.instance, s.P.logger.With("service", s.Name()))
}

// Shutdown implements the Shutdowner interface.
// It delegates to the underlying service instance if it implements the Shutdowner interface.
func (s *ServiceSingleton[I, S]) Shutdown(ctx context.Context) error {
	return shutdownService(ctx, s.instance, s.P.logger.With("service", s.Name()))
}

// Make creates a new instance of the service without initializing it.
// This is used during dependency graph construction to analyze the service's dependencies.
func (s *ServiceSingleton[I, S]) Make() any {
	return new(S)
}

// Instance returns the stored instance of the service.
// For singleton services, this always returns the same instance after initialization.
func (s *ServiceSingleton[I, S]) Instance(_ context.Context) (any, error) {
	return s.instance, nil
}

// BeforeInit registers a hook function that will be called before the service is initialized.
// This can be used to customize the service instance before its Init method is called.
func (s *ServiceSingleton[I, S]) BeforeInit(hook LifecycleHook[*S]) *ServiceSingleton[I, S] {
	s.beforeInit = hook
	return s
}

// Validate implements the ServiceDef interface.
// It validates that the service implementation satisfies all required interfaces and constraints.
func (s *ServiceSingleton[I, S]) Validate(ctx context.Context) error {
	return validateService[I, S](ctx)
}

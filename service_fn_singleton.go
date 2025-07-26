package pal

import (
	"context"
)

// ServiceFnSingleton is a singleton service that is created using a function.
// It is created during initialization and reused for the lifetime of the application.
type ServiceFnSingleton[T any] struct {
	P  *Pal
	fn func(ctx context.Context) (T, error)

	hooks LifecycleHooks[T]

	instance T
}

func (c *ServiceFnSingleton[T]) Dependencies() []ServiceDef {
	return nil
}

// Run executes the service if it implements the Runner interface.
func (c *ServiceFnSingleton[T]) Run(ctx context.Context) error {
	return runService(ctx, c.instance, c.P.logger.With("service", c.Name()))
}

// Init initializes the service by calling the provided function to create the instance.
func (c *ServiceFnSingleton[T]) Init(ctx context.Context) error {
	instance, err := c.fn(ctx)
	if err != nil {
		return err
	}

	c.instance = instance

	return nil
}

func (c *ServiceFnSingleton[T]) RunConfig() *RunConfig {
	return nil
}

// HealthCheck performs a health check on the service if it implements the HealthChecker interface.
func (c *ServiceFnSingleton[T]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, c.instance, c.hooks.HealthCheck, c.P.logger.With("service", c.Name()))
}

// Shutdown gracefully shuts down the service if it implements the Shutdowner interface.
func (c *ServiceFnSingleton[T]) Shutdown(ctx context.Context) error {
	return shutdownService(ctx, c.instance, c.hooks.Shutdown, c.P.logger.With("service", c.Name()))
}

// Make returns nil for singleton services as they are created during initialization.
func (c *ServiceFnSingleton[T]) Make() any {
	return nil
}

// Instance returns the singleton instance of the service.
func (c *ServiceFnSingleton[T]) Instance(_ context.Context) (any, error) {
	return c.instance, nil
}

func (c *ServiceFnSingleton[T]) ToShutdown(hook LifecycleHook[T]) *ServiceFnSingleton[T] {
	c.hooks.Shutdown = hook
	return c
}

// ToHealthCheck registers a hook function that will be called to perform a health check on the service.
// If the service implements the HealthChecker interface, the HealthCheck() method is not called,
// the hook has higher priority.
func (c *ServiceFnSingleton[T]) ToHealthCheck(hook LifecycleHook[T]) *ServiceFnSingleton[T] {
	c.hooks.HealthCheck = hook
	return c
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceFnSingleton[T]) Name() string {
	return elem[T]().String()
}

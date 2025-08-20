package pal

import (
	"context"
)

// ServiceConst is a service that wraps a constant value.
// It is used to register existing instances as services.
type ServiceConst[T any] struct {
	ServiceTyped[T]

	hooks LifecycleHooks[T]

	instance T
}

func (c *ServiceConst[T]) RunConfig() *RunConfig {
	return runConfig(c.instance)
}

// Run executes the service if it implements the Runner interface.
func (c *ServiceConst[T]) Run(ctx context.Context) error {
	return runService(ctx, c.Name(), c.instance, c.P)
}

// Init is a no-op for const services as they are already initialized.
// It injects dependencies to the stored instance and calls its Init method if it implements Initer.
func (c *ServiceConst[T]) Init(ctx context.Context) error {
	return initService(ctx, c.Name(), c.instance, c.hooks.Init, c.P)
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceConst[T]) Make() any {
	return c.instance
}

// HealthCheck performs a health check on the service if it implements the HealthChecker interface.
func (c *ServiceConst[T]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, c.Name(), c.instance, c.hooks.HealthCheck, c.P)
}

// Shutdown gracefully shuts down the service if it implements the Shutdowner interface.
func (c *ServiceConst[T]) Shutdown(ctx context.Context) error {
	return shutdownService(ctx, c.Name(), c.instance, c.hooks.Shutdown, c.P)
}

// Instance returns the constant instance of the service.
func (c *ServiceConst[T]) Instance(_ context.Context, _ ...any) (any, error) {
	return c.instance, nil
}

// ToInit registers a hook function that will be called to initialize the service.
// This hook is called after the service is injected with its dependencies.
// If the service implements the Initer interface, the Init() method is not called,
// the hook has higher priority.
func (c *ServiceConst[T]) ToInit(hook LifecycleHook[T]) *ServiceConst[T] {
	c.hooks.Init = hook
	return c
}

// ToShutdown registers a hook function that will be called to shutdown the service.
// This hook is called before service's dependencies are shutdown.
// If the service implements the Shutdowner interface, the Shutdown() method is not called,
// the hook has higher priority.
func (c *ServiceConst[T]) ToShutdown(hook LifecycleHook[T]) *ServiceConst[T] {
	c.hooks.Shutdown = hook
	return c
}

// ToHealthCheck registers a hook function that will be called to perform a health check on the service.
// If the service implements the HealthChecker interface, the HealthCheck() method is not called,
// the hook has higher priority.
func (c *ServiceConst[T]) ToHealthCheck(hook LifecycleHook[T]) *ServiceConst[T] {
	c.hooks.HealthCheck = hook
	return c
}

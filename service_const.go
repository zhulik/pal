package pal

import (
	"context"
)

// ServiceConst is a service that wraps a constant value.
// It is used to register existing instances as services.
type ServiceConst[T any] struct {
	P *Pal

	beforeShutdown LifecycleHook[T]

	instance T
}

// Run executes the service if it implements the Runner interface.
func (c *ServiceConst[T]) Run(ctx context.Context) error {
	return runService(ctx, c.instance, c.P.logger.With("service", c.Name()))
}

// Init is a no-op for const services as they are already initialized.
func (c *ServiceConst[T]) Init(_ context.Context) error {
	return nil
}

// HealthCheck performs a health check on the service if it implements the HealthChecker interface.
func (c *ServiceConst[T]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, c.instance, c.P.logger.With("service", c.Name()))
}

// Shutdown gracefully shuts down the service if it implements the Shutdowner interface.
func (c *ServiceConst[T]) Shutdown(ctx context.Context) error {
	if c.beforeShutdown != nil {
		c.P.logger.Info("Calling BeforeShutdown hook")
		err := c.beforeShutdown(ctx, c.instance)
		if err != nil {
			c.P.logger.Info("BeforeShutdown failed", "error", err)
			return err
		}
	}
	return shutdownService(ctx, c.instance, c.P.logger.With("service", c.Name()))
}

// Make returns nil for const services as they are already created.
func (c *ServiceConst[T]) Make() any {
	return nil
}

// Instance returns the constant instance of the service.
func (c *ServiceConst[T]) Instance(_ context.Context) (any, error) {
	return c.instance, nil
}

func (c *ServiceConst[T]) BeforeShutdown(hook LifecycleHook[T]) *ServiceConst[T] {
	c.beforeShutdown = hook
	return c
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceConst[T]) Name() string {
	return elem[T]().String()
}

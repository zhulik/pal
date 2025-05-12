package pal

import (
	"context"
)

// ServiceConst is a service that wraps a constant value.
// It is used to register existing instances as services.
type ServiceConst[T any] struct {
	P *Pal

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

// Name returns the name of the service, which is the type name of T.
func (c *ServiceConst[T]) Name() string {
	return elem[T]().String()
}

// Validate performs validation of the service configuration.
func (c *ServiceConst[T]) Validate(_ context.Context) error {
	return nil
}

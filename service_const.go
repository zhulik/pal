package pal

import (
	"context"
)

// ServiceConst is a service that wraps a constant value.
// It is used to register existing instances as services.
type ServiceConst[T any] struct {
	P *Pal

	hooks LifecycleHooks[T]

	instance T
}

func (c *ServiceConst[T]) RunConfig() *RunConfig {
	return runConfig(c.instance)
}

func (c *ServiceConst[T]) Dependencies() []ServiceDef {
	return nil
}

// Run executes the service if it implements the Runner interface.
func (c *ServiceConst[T]) Run(ctx context.Context) error {
	return runService(ctx, c.instance, c.P.logger.With("service", c.Name()))
}

// Init is a no-op for const services as they are already initialized.
// It injects dependencies to the stored instance and calls its Init method if it implements Initer.
func (c *ServiceConst[T]) Init(ctx context.Context) error {
	logger := c.P.logger.With("service", c.Name())

	err := c.P.InjectInto(ctx, c.instance)
	if err != nil {
		return err
	}

	if c.hooks.Init != nil {
		logger.Debug("Calling BeforeInit hook")
		err := c.hooks.Init(ctx, c.instance)
		if err != nil {
			logger.Error("BeforeInit hook failed", "error", err)
			return err
		}
	}

	if initer, ok := any(c.instance).(Initer); ok && any(c.instance) != any(c.P) {
		logger.Debug("Calling Init method")
		if err := initer.Init(ctx); err != nil {
			logger.Error("Init failed", "error", err)
			return err
		}
	}
	return nil
}

// HealthCheck performs a health check on the service if it implements the HealthChecker interface.
func (c *ServiceConst[T]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, c.instance, c.P.logger.With("service", c.Name()))
}

// Shutdown gracefully shuts down the service if it implements the Shutdowner interface.
func (c *ServiceConst[T]) Shutdown(ctx context.Context) error {
	if c.hooks.Shutdown != nil {
		c.P.logger.Debug("Calling BeforeShutdown hook")
		err := c.hooks.Shutdown(ctx, c.instance)
		if err != nil {
			c.P.logger.Error("BeforeShutdown failed", "error", err)
			return err
		}
	}
	return shutdownService(ctx, c.instance, c.P.logger.With("service", c.Name()))
}

// Make returns the stored instance so pal knows the entire dependency tree.
func (c *ServiceConst[T]) Make() any {
	return c.instance
}

// Instance returns the constant instance of the service.
func (c *ServiceConst[T]) Instance(_ context.Context) (any, error) {
	return c.instance, nil
}

// BeforeInit registers a hook function that will be called before the service is initialized.
// This can be used to customize the service instance before its Init method is called.
func (c *ServiceConst[T]) BeforeInit(hook LifecycleHook[T]) *ServiceConst[T] {
	c.hooks.Init = hook
	return c
}

func (c *ServiceConst[T]) BeforeShutdown(hook LifecycleHook[T]) *ServiceConst[T] {
	c.hooks.Shutdown = hook
	return c
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceConst[T]) Name() string {
	return elem[T]().String()
}

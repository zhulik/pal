package pal

import (
	"context"
)

type ServiceFnSingleton[T any] struct {
	fn func(ctx context.Context) (T, error)

	instance T
}

func (c *ServiceFnSingleton[T]) Run(ctx context.Context) error {
	return runService(ctx, c.instance, c.Name())
}

func (c *ServiceFnSingleton[T]) Init(ctx context.Context) error {
	instance, err := c.fn(ctx)
	if err != nil {
		return err
	}

	c.instance = instance

	return nil
}

func (c *ServiceFnSingleton[T]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, c.instance)
}

func (c *ServiceFnSingleton[T]) Shutdown(ctx context.Context) error {
	return shutdownService(ctx, c.instance)
}

func (c *ServiceFnSingleton[T]) Make() any {
	return nil
}

func (c *ServiceFnSingleton[T]) Instance(_ context.Context) (any, error) {
	return c.instance, nil
}

func (c *ServiceFnSingleton[T]) Name() string {
	return elem[T]().String()
}

func (c *ServiceFnSingleton[T]) Validate(_ context.Context) error {
	return nil
}

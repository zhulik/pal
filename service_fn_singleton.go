package pal

import (
	"context"
)

type ServiceFnSingleton[I any, S any] struct {
	fn func(ctx context.Context) (*S, error)

	instance I
}

func (c *ServiceFnSingleton[I, S]) Run(ctx context.Context) error {
	return runService(ctx, c.instance, c.Name())
}

func (c *ServiceFnSingleton[I, S]) Init(ctx context.Context) error {
	instance, err := c.fn(ctx)
	if err != nil {
		return err
	}

	c.instance = any(instance).(I)

	return nil
}

func (c *ServiceFnSingleton[I, S]) HealthCheck(ctx context.Context) error {
	return healthcheckService(ctx, c.instance)
}

func (c *ServiceFnSingleton[I, S]) Shutdown(ctx context.Context) error {
	return shutdownService(ctx, c.instance)
}

func (c *ServiceFnSingleton[I, S]) Make() any {
	return nil
}

func (c *ServiceFnSingleton[I, S]) Instance(_ context.Context) (any, error) {
	return c.instance, nil
}

func (c *ServiceFnSingleton[I, S]) Name() string {
	return elem[I]().String()
}

func (c *ServiceFnSingleton[I, S]) Validate(_ context.Context) error {
	return nil
}

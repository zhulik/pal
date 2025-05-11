package pal

import (
	"context"
)

type ServiceConst[T any] struct {
	instance T
}

func (c *ServiceConst[T]) Run(ctx context.Context) error {
	return runService(ctx, c.instance, c.Name())
}

func (c *ServiceConst[T]) Init(_ context.Context) error {
	return nil
}

func (c *ServiceConst[T]) HealthCheck(ctx context.Context) error {
	if h, ok := any(c.instance).(HealthChecker); ok {
		return h.HealthCheck(ctx)
	}
	return nil
}

func (c *ServiceConst[T]) Shutdown(ctx context.Context) error {
	if h, ok := any(c.instance).(Shutdowner); ok {
		return h.Shutdown(ctx)
	}
	return nil
}

func (c *ServiceConst[T]) Make() any {
	return nil
}

func (c *ServiceConst[T]) Instance(_ context.Context) (any, error) {
	return c.instance, nil
}

func (c *ServiceConst[T]) Name() string {
	return elem[T]().String()
}

func (c *ServiceConst[T]) Validate(_ context.Context) error {
	return nil
}

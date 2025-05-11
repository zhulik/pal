package pal

import (
	"context"
)

type ConstService[T any] struct {
	instance T
}

func (c *ConstService[T]) Init(_ context.Context) error {
	return nil
}

func (c *ConstService[T]) HealthCheck(ctx context.Context) error {
	if h, ok := any(c.instance).(HealthChecker); ok {
		return h.HealthCheck(ctx)
	}
	return nil
}

func (c *ConstService[T]) Shutdown(ctx context.Context) error {
	if h, ok := any(c.instance).(Shutdowner); ok {
		return h.Shutdown(ctx)
	}
	return nil
}

func (c *ConstService[T]) Make() any {
	return nil
}

func (c *ConstService[T]) Instance(_ context.Context) (any, error) {
	return c.instance, nil
}

func (c *ConstService[T]) Name() string {
	return elem[T]().String()
}

func (c *ConstService[T]) IsSingleton() bool {
	return true
}

func (c *ConstService[T]) IsRunner() bool {
	_, runner := any(c.instance).(Runner)
	return runner
}

func (c *ConstService[T]) Validate(_ context.Context) error {
	return nil
}

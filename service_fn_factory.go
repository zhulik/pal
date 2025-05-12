package pal

import (
	"context"
)

type ServiceFnFactory[T any] struct {
	fn func(ctx context.Context) (T, error)
}

func (c *ServiceFnFactory[T]) Run(_ context.Context) error {
	return nil
}

func (c *ServiceFnFactory[T]) Init(_ context.Context) error {
	return nil
}

func (c *ServiceFnFactory[T]) HealthCheck(_ context.Context) error {
	return nil
}

func (c *ServiceFnFactory[T]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceFnFactory[T]) Make() any {
	return nil
}

func (c *ServiceFnFactory[T]) Instance(ctx context.Context) (any, error) {
	return c.fn(ctx)
}

func (c *ServiceFnFactory[T]) Name() string {
	return elem[T]().String()
}

func (c *ServiceFnFactory[T]) Validate(_ context.Context) error {
	return nil
}

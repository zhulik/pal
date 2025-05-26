package pal

import (
	"context"
)

// ServiceFnFactory is a factory service that creates a new instance each time it is invoked.
// It uses the provided function to create the instance.
type ServiceFnFactory[T any] struct {
	P  *Pal
	fn func(ctx context.Context) (T, error)
}

func (c *ServiceFnFactory[T]) Dependencies() []ServiceDef {
	return nil
}

// Run is a no-op for factory services as they don't run in the background.
func (c *ServiceFnFactory[T]) Run(_ context.Context) error {
	return nil
}

// Init is a no-op for factory services as they are created on demand.
func (c *ServiceFnFactory[T]) Init(_ context.Context) error {
	return nil
}

// HealthCheck is a no-op for factory services as they are created on demand.
func (c *ServiceFnFactory[T]) HealthCheck(_ context.Context) error {
	return nil
}

// Shutdown is a no-op for factory services as they are created on demand.
func (c *ServiceFnFactory[T]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceFnFactory[T]) RunConfig() *RunConfig {
	return nil
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceFnFactory[T]) Make() any {
	return nil
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFnFactory[T]) Instance(ctx context.Context) (any, error) {
	return c.fn(ctx)
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceFnFactory[T]) Name() string {
	return elem[T]().String()
}

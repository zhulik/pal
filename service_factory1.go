package pal

import (
	"context"
	"fmt"
)

// ServiceFactory1 is a factory service that creates a new instance each time it is invoked.
// It uses the provided function to create the instance.
type ServiceFactory1[T any, P1 any] struct {
	P  *Pal
	fn func(ctx context.Context, p1 P1) (T, error)
}

func (c *ServiceFactory1[T, P1]) Dependencies() []ServiceDef {
	return nil
}

// Run is a no-op for factory services as they don't run in the background.
func (c *ServiceFactory1[T, P1]) Run(_ context.Context) error {
	return nil
}

// Init is a no-op for factory services as they are created on demand.
func (c *ServiceFactory1[T, P1]) Init(_ context.Context) error {
	return nil
}

// HealthCheck is a no-op for factory services as they are created on demand.
func (c *ServiceFactory1[T, P1]) HealthCheck(_ context.Context) error {
	return nil
}

// Shutdown is a no-op for factory services as they are created on demand.
func (c *ServiceFactory1[T, P1]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceFactory1[T, P1]) RunConfig() *RunConfig {
	return nil
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceFactory1[T, P1]) Make() any {
	var t T
	return t
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFactory1[T, P1]) Instance(ctx context.Context, args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %d, expected 1", ErrServiceInvalidArgumentsCount, len(args))
	}

	p1, ok := args[0].(P1)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[0], p1)
	}

	return c.fn(ctx, p1)
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceFactory1[T, P1]) Name() string {
	return elem[T]().String()
}

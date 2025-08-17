package pal

import (
	"context"
	"fmt"
)

// ServiceFactory2 is a factory service that creates a new instance each time it is invoked.
// It uses the provided function to create the instance.
type ServiceFactory2[T any, P1 any, P2 any] struct {
	P  *Pal
	fn func(ctx context.Context, p1 P1, p2 P2) (T, error)
}

func (c *ServiceFactory2[T, P1, P2]) Dependencies() []ServiceDef {
	return nil
}

// Run is a no-op for factory services as they don't run in the background.
func (c *ServiceFactory2[T, P1, P2]) Run(_ context.Context) error {
	return nil
}

// Init is a no-op for factory services as they are created on demand.
func (c *ServiceFactory2[T, P1, P2]) Init(_ context.Context) error {
	return nil
}

// HealthCheck is a no-op for factory services as they are created on demand.
func (c *ServiceFactory2[T, P1, P2]) HealthCheck(_ context.Context) error {
	return nil
}

// Shutdown is a no-op for factory services as they are created on demand.
func (c *ServiceFactory2[T, P1, P2]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceFactory2[T, P1, P2]) RunConfig() *RunConfig {
	return nil
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceFactory2[T, P1, P2]) Make() any {
	var t T
	return t
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFactory2[T, P1, P2]) Instance(ctx context.Context, args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: %d, expected 2", ErrServiceInvalidArgumentsCount, len(args))
	}

	p1, ok := args[0].(P1)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[0], p1)
	}

	p2, ok := args[1].(P2)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[1], p2)
	}

	return c.fn(ctx, p1, p2)
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceFactory2[T, P1, P2]) Name() string {
	return elem[T]().String()
}

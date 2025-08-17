package pal

import (
	"context"
	"fmt"
)

// ServiceFactory5 is a factory service that creates a new instance each time it is invoked.
// It uses the provided function to create the instance.
type ServiceFactory5[T any, P1 any, P2 any, P3 any, P4 any, P5 any] struct {
	P  *Pal
	fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (T, error)
}

func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) Dependencies() []ServiceDef {
	return nil
}

// Run is a no-op for factory services as they don't run in the background.
func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) Run(_ context.Context) error {
	return nil
}

// Init is a no-op for factory services as they are created on demand.
func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) Init(_ context.Context) error {
	return nil
}

// HealthCheck is a no-op for factory services as they are created on demand.
func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) HealthCheck(_ context.Context) error {
	return nil
}

// Shutdown is a no-op for factory services as they are created on demand.
func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) RunConfig() *RunConfig {
	return nil
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) Make() any {
	var t T
	return t
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) Instance(ctx context.Context, args ...any) (any, error) {
	if len(args) != 5 {
		return nil, fmt.Errorf("%w: %d, expected 5", ErrServiceInvalidArgumentsCount, len(args))
	}

	p1, ok := args[0].(P1)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[0], p1)
	}

	p2, ok := args[1].(P2)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[1], p2)
	}

	p3, ok := args[2].(P3)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[2], p3)
	}

	p4, ok := args[3].(P4)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[3], p4)
	}

	p5, ok := args[4].(P5)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[4], p5)
	}

	return c.fn(ctx, p1, p2, p3, p4, p5)
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceFactory5[T, P1, P2, P3, P4, P5]) Name() string {
	return elem[T]().String()
}

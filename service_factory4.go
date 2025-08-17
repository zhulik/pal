package pal

import (
	"context"
	"fmt"
)

// ServiceFactory4 is a factory service that creates a new instance each time it is invoked.
// It uses the provided function with four arguments to create the instance.
type ServiceFactory4[T any, P1 any, P2 any, P3 any, P4 any] struct {
	ServiceTyped[T]
	fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4) (T, error)
}

func (c *ServiceFactory4[T, P1, P2, P3, P4]) Arguments() int {
	return 4
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFactory4[T, P1, P2, P3, P4]) Instance(ctx context.Context, args ...any) (any, error) {
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

	return c.fn(ctx, p1, p2, p3, p4)
}

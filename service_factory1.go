package pal

import (
	"context"
	"fmt"
)

// ServiceFactory1 is a factory service that creates a new instance each time it is invoked.
// It uses the provided function to create the instance.
type ServiceFactory1[T any, P1 any] struct {
	ServiceTyped[T]
	fn func(ctx context.Context, p1 P1) (T, error)
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

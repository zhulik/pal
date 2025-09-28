package pal

import (
	"context"
	"fmt"
)

// ServiceFactory1 is a factory service that creates a new instance each time it is invoked.
// It uses the provided function with one argument to create the instance.
type ServiceFactory1[I any, T any, P1 any] struct {
	ServiceFactory[I, T]
	fn func(ctx context.Context, p1 P1) (T, error)
}

func (c *ServiceFactory1[I, T, P1]) Arguments() int {
	return 1
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFactory1[I, T, P1]) Instance(ctx context.Context, args ...any) (any, error) {
	p1, ok := args[0].(P1)
	if !ok {
		return nil, fmt.Errorf("%w: %T, expected %T", ErrServiceInvalidArgumentType, args[0], p1)
	}

	instance, err := c.fn(ctx, p1)
	if err != nil {
		return nil, err
	}

	err = initService(ctx, c.Name(), instance, nil, c.P)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// Factory returns a function that creates a new instance of the service.
func (c *ServiceFactory1[I, T, P1]) Factory() any {
	return func(ctx context.Context, p1 P1) (I, error) {
		instance, err := c.Instance(ctx, p1)
		if err != nil {
			var i I
			return i, err
		}
		return instance.(I), nil
	}
}

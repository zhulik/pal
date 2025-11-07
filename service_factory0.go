package pal

import (
	"context"
)

// ServiceFactory0 is a factory service that creates a new instance each time it is invoked.
// It uses the provided function with no arguments to create the instance.
type ServiceFactory0[I any, T any] struct {
	ServiceFactory[I, T]
	fn func(ctx context.Context) (T, error)
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFactory0[I, T]) Instance(ctx context.Context, _ ...any) (any, error) {
	instance, err := c.fn(ctx)
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
// The returned function has the signature func(ctx context.Context) (I, error).
func (c *ServiceFactory0[I, T]) Factory() any {
	return func(ctx context.Context) (I, error) {
		instance, err := c.Instance(ctx)
		if err != nil {
			var i I
			return i, err
		}
		return instance.(I), nil
	}
}

// MustFactory returns a function that creates a new instance of the service.
// The returned function has the signature func(ctx context.Context) I.
// If the instance creation fails, it panics.
func (c *ServiceFactory0[I, T]) MustFactory() any {
	return func(ctx context.Context) I {
		return must(c.Instance(ctx)).(I)
	}
}

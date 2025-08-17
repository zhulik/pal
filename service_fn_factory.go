package pal

import (
	"context"
)

// ServiceFnFactory is a factory service that creates a new instance each time it is invoked.
// It uses the provided function to create the instance.
type ServiceFnFactory[T any] struct {
	ServiceTyped[T]
	fn func(ctx context.Context) (T, error)
}

// Instance creates and returns a new instance of the service using the provided function.
func (c *ServiceFnFactory[T]) Instance(ctx context.Context, _ ...any) (any, error) {
	return c.fn(ctx)
}

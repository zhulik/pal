package pal

import "context"

type ServiceTyped[T any] struct {
	P *Pal
}

func (c *ServiceTyped[T]) Dependencies() []ServiceDef {
	return nil
}

// Run is a no-op for factory services as they don't run in the background.
func (c *ServiceTyped[T]) Run(_ context.Context) error {
	return nil
}

// Init is a no-op for factory services as they are created on demand.
func (c *ServiceTyped[T]) Init(_ context.Context) error {
	return nil
}

// HealthCheck is a no-op for factory services as they are created on demand.
func (c *ServiceTyped[T]) HealthCheck(_ context.Context) error {
	return nil
}

// Shutdown is a no-op for factory services as they are created on demand.
func (c *ServiceTyped[T]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceTyped[T]) RunConfig() *RunConfig {
	return nil
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceTyped[T]) Make() any {
	var t T
	return t
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceTyped[T]) Name() string {
	return elem[T]().String()
}

func (c *ServiceTyped[T]) Arguments() int {
	return 0
}

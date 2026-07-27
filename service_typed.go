package pal

type ServiceTyped[T any] struct {
	P    *Pal
	name string
}

func (c *ServiceTyped[T]) Dependencies() []ServiceDef {
	return nil
}

func (c *ServiceTyped[T]) ShouldWaitForRunner() *bool {
	return nil
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceTyped[T]) Make() any {
	var t T
	return t
}

// Name returns the name of the service, which is the type name of T.
func (c *ServiceTyped[T]) Name() string {
	return c.name
}

func (c *ServiceTyped[T]) Arguments() int {
	return 0
}

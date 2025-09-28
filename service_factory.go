package pal

import "reflect"

type ServiceFactory[I any, T any] struct {
	ServiceTyped[I]
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceFactory[I, T]) Make() any {
	var t T
	typ := reflect.TypeOf(t).Elem()
	return reflect.New(typ).Interface().(I)
}

package pal

import "reflect"

// ServiceFactory is the shared base for arity-specific factory wrappers.
//
// Advanced: prefer [ProvideFactory0]–[ProvideFactory5]; this type remains exported for power users.
type ServiceFactory[I any, T any] struct {
	ServiceTyped[I]
}

// Make is a no-op for factory services as they are created on demand.
func (c *ServiceFactory[I, T]) Make() any {
	var t T
	typ := reflect.TypeOf(t).Elem()
	return reflect.New(typ).Interface().(I)
}

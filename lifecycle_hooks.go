package pal

import "context"

// LifecycleHook is a function type that can be registered to run at specific points in a service's lifecycle.
// It receives the service instance and a contex, and can return an error to indicate failure.
// These hooks are typically used with ToInit methods to customize service initialization.
type LifecycleHook[T any] func(ctx context.Context, service T) error

// LifecycleHooks is a collection of hooks that can be registered to run at specific points in a service's lifecycle.
type LifecycleHooks[T any] struct {
	Init        LifecycleHook[T]
	Shutdown    LifecycleHook[T]
	HealthCheck LifecycleHook[T]
}

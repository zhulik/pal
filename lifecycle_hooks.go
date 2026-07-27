package pal

import "context"

// LifecycleHook is a function type that can be registered to run at specific points in a service's lifecycle.
// It receives the service instance, a context, and an [Invoker] (typically the running [Pal]), and can return an error to indicate failure.
// These hooks are typically used with ToInit methods to customize service initialization.
type LifecycleHook[T any] func(ctx context.Context, service T, invoker Invoker) error

// lifecycleHooks is a collection of hooks that can be registered to run at specific points in a service's lifecycle.
type lifecycleHooks[T any] struct {
	Init        LifecycleHook[T]
	Shutdown    LifecycleHook[T]
	HealthCheck LifecycleHook[T]
}

// Hookable is the fluent registration surface returned by [Provide], [ProvideNamed], [ProvideFn], and [ProvideNamedFn].
// Concrete wrappers such as [ServiceConst] and [ServiceFnSingleton] remain exported for advanced use.
type Hookable[T any] interface {
	ServiceDef
	ToInit(hook LifecycleHook[T]) Hookable[T]
	ToShutdown(hook LifecycleHook[T]) Hookable[T]
	ToHealthCheck(hook LifecycleHook[T]) Hookable[T]
}

package pal

import "context"

// LifecycleHook is a function type that can be registered to run at specific points in a service's lifecycle.
// It receives the service instance and a contex, and can return an error to indicate failure.
// These hooks are typically used with BeforeInit methods to customize service initialization.
type LifecycleHook[T any] func(ctx context.Context, service T) error

package pal

import (
	"context"
	"fmt"
)

// Provide registers a singleton service with pal. `I` is the type `S` will be cast to. It's also used to generate
// service name. Typically, `I` would be one of:
// - An interface, in this case S must implement it. Used when I may have multiple implementations like mocks for tests.
// - A pointer to `S`. For instance,`Provide[*Foo, Foo]()`. Used when mocking is not required.
// Only one instance of the service will be created and reused.
func Provide[I any, S any]() *ServiceSingleton[I, S] {
	return &ServiceSingleton[I, S]{}
}

// ProvideFn registers a singleton that is build with a given function.
func ProvideFn[T any](fn func(ctx context.Context) (T, error)) *ServiceFnSingleton[T] {
	return &ServiceFnSingleton[T]{
		fn: fn,
	}
}

// ProvideFactory registers a factory service with pal. See Provide for info on type arguments.
// A new factory service instance is created every time the service is invoked.
// it's the caller's responsibility to shut down the service, pal will also not healthcheck it.
func ProvideFactory[I any, S any]() *ServiceFactory[I, S] {
	return &ServiceFactory[I, S]{}
}

// ProvideFnFactory registers a factory service that is build with a given function.
func ProvideFnFactory[T any](fn func(ctx context.Context) (T, error)) *ServiceFnFactory[T] {
	return &ServiceFnFactory[T]{
		fn: fn,
	}
}

// ProvideRunner turns the given function into a runner. It will run in the background, and the passed context will
// be cancelled on app shutdown.
func ProvideRunner(fn func(ctx context.Context) error) *ServiceRunner {
	return &ServiceRunner{fn}
}

// ProvideConst registers a const as a service.
func ProvideConst[T any](value T) *ServiceConst[T] {
	return &ServiceConst[T]{instance: value}
}

// Invoke retrieves or creates an instance of type I from the given Pal container.
func Invoke[T any](ctx context.Context, invoker Invoker) (T, error) {
	name := elem[T]().String()

	a, err := invoker.Invoke(ctx, name)
	if err != nil {
		return empty[T](), err
	}

	casted, ok := a.(T)
	if !ok {
		return empty[T](), fmt.Errorf("%w: %s. %+v does not implement %s", ErrServiceInvalid, name, a, name)
	}

	return casted, nil
}

// Build resolves dependencies for a struct of type S using the provided context and Pal instance.
// It initializes the struct's fields by injecting appropriate dependencies based on the field types.
// Returns the fully initialized struct or an error if dependency resolution fails.
func Build[T any](ctx context.Context, invoker Invoker) (*T, error) {
	s := new(T)

	err := InjectInto[T](ctx, invoker, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// InjectInto populates the fields of a struct of type T with dependencies obtained from the given Invoker.
// It only sets fields that are exported and match a resolvable dependency, skipping fields when ErrServiceNotFound occurs.
// Returns an error if dependency invocation fails or other unrecoverable errors occur during injection.
func InjectInto[T any](ctx context.Context, invoker Invoker, s *T) error {
	return invoker.InjectInto(ctx, s)
}

package pal

import (
	"context"
	"fmt"
	"reflect"
)

// Provide registers a const as a service. `T` is used to generating service name.
// Typically, `T` would be one of:
// - An interface, in this case passed value must implement it. Used when T may have multiple implementations like mocks for tests.
// - A pointer to an instance of `T`. For instance,`Provide[*Foo](&Foo{})`. Used when mocking is not required.
// If the passed value implements Initer, Init() will be called.
func Provide[T any](value T) *ServiceConst[T] {
	validatePointerToStruct(value)

	return &ServiceConst[T]{instance: value}
}

// ProvideFn registers a singleton built with a given function.
func ProvideFn[T any](fn func(ctx context.Context) (T, error)) *ServiceFnSingleton[T] {
	return &ServiceFnSingleton[T]{
		fn: fn,
	}
}

// ProvideFactory registers a factory service with pal. See Provide for info on type arguments.
// A new factory service instance is created every time the service is invoked.
// it's the caller's responsibility to shut down the service, pal will also not healthcheck it.
func ProvideFactory[T any](value T) *ServiceFactory[T] {
	validatePointerToStruct(value)

	return &ServiceFactory[T]{
		referenceInstance: value,
	}
}

// ProvideFnFactory registers a factory service that is build with a given function.
func ProvideFnFactory[T any](fn func(ctx context.Context) (T, error)) *ServiceFnFactory[T] {
	return &ServiceFnFactory[T]{
		fn: fn,
	}
}

// ProvideRunner turns the given function into a runner. It will run in the background, and the passed context will
// be canceled on app shutdown.
func ProvideRunner(fn func(ctx context.Context) error) *ServiceRunner {
	return &ServiceRunner{fn}
}

// ProvideList registers a list of given services.
func ProvideList(services ...ServiceDef) *ServiceList {
	return &ServiceList{services}
}

// ProvidePal registers all services for the given pal instance
func ProvidePal(pal *Pal) *ServiceList {
	services := make([]ServiceDef, 0, len(pal.Services()))
	for _, v := range pal.Services() {
		if v.Name() != "*pal.Pal" {
			services = append(services, v)
		}
	}

	return ProvideList(services...)
}

// Invoke retrieves or creates an instance of type T from the given Pal container.
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

// MustInvoke is like Invoke but panics if an error occurs.
func MustInvoke[T any](ctx context.Context, invoker Invoker) T {
	return must(Invoke[T](ctx, invoker))
}

// Build resolves dependencies for a struct of type T using the provided context and Invoker.
// It initializes the struct's fields by injecting appropriate dependencies based on the field types.
// Returns the fully initialized struct or an error if dependency resolution fails.
func Build[T any](ctx context.Context, invoker Invoker) (*T, error) {
	s := new(T)

	err := InjectInto(ctx, invoker, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// MustBuild is like Build but panics if an error occurs.
func MustBuild[T any](ctx context.Context, invoker Invoker) *T {
	return must(Build[T](ctx, invoker))
}

// InjectInto populates the fields of a struct of type T with dependencies obtained from the given Invoker.
// It only sets fields that are exported and match a resolvable dependency, skipping fields when ErrServiceNotFound occurs.
// Returns an error if dependency invocation fails or other unrecoverable errors occur during injection.
func InjectInto[T any](ctx context.Context, invoker Invoker, s *T) error {
	return invoker.InjectInto(ctx, s)
}

// MustInjectInto is like InjectInto but panics if an error occurs.
func MustInjectInto[T any](ctx context.Context, invoker Invoker, s *T) {
	must("", InjectInto(ctx, invoker, s))
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func validatePointerToStruct(value any) {
	val := reflect.ValueOf(value)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		panic(fmt.Sprintf("Argument must be a non-nil pointer to a struct, got %T", value))
	}
}

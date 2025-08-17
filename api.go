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

	return &ServiceFactory[T]{referenceInstance: value}
}

// ProvideFactory0 registers a factory service that is build with a given function with no arguments.
func ProvideFactory0[T any](fn func(ctx context.Context) (T, error)) *ServiceFactory0[T] {
	return &ServiceFactory0[T]{
		fn: fn,
	}
}

// ProvideRunner turns the given function into a runner. It will run in the background, and the passed context will
// be canceled on app shutdown.
func ProvideRunner(fn func(ctx context.Context) error) *ServiceRunner {
	return &ServiceRunner{fn: fn}
}

// ProvideList registers a list of given services.
func ProvideList(services ...ServiceDef) *ServiceList {
	return &ServiceList{Services: services}
}

// ProvideFactory1 registers a factory service that is built in runtime with a given function that takes one argument.
func ProvideFactory1[T any, P1 any](fn func(ctx context.Context, p1 P1) (T, error)) *ServiceFactory1[T, P1] {
	return &ServiceFactory1[T, P1]{fn: fn}
}

// ProvideFactory2 registers a factory service that is built in runtime with a given function that takes two arguments.
func ProvideFactory2[T any, P1 any, P2 any](fn func(ctx context.Context, p1 P1, p2 P2) (T, error)) *ServiceFactory2[T, P1, P2] {
	return &ServiceFactory2[T, P1, P2]{fn: fn}
}

// ProvideFactory3 registers a factory service that is built in runtime with a given function that takes three arguments.
func ProvideFactory3[T any, P1 any, P2 any, P3 any](fn func(ctx context.Context, p1 P1, p2 P2, p3 P3) (T, error)) *ServiceFactory3[T, P1, P2, P3] {
	return &ServiceFactory3[T, P1, P2, P3]{fn: fn}
}

// ProvideFactory4 registers a factory service that is built in runtime with a given function that takes four arguments.
func ProvideFactory4[T any, P1 any, P2 any, P3 any, P4 any](fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4) (T, error)) *ServiceFactory4[T, P1, P2, P3, P4] {
	return &ServiceFactory4[T, P1, P2, P3, P4]{fn: fn}
}

// ProvideFactory5 registers a factory service that is built in runtime with a given function that takes five arguments.
func ProvideFactory5[T any, P1 any, P2 any, P3 any, P4 any, P5 any](fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (T, error)) *ServiceFactory5[T, P1, P2, P3, P4, P5] {
	return &ServiceFactory5[T, P1, P2, P3, P4, P5]{fn: fn}
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
func Invoke[T any](ctx context.Context, invoker Invoker, args ...any) (T, error) {
	name := elem[T]().String()

	a, err := invoker.Invoke(ctx, name, args...)
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

// MustBuild is like Build but panics if an error occurs.x
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

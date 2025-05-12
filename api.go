package pal

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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
func ProvideFn[I any, S any](fn func(ctx context.Context) (*S, error)) *ServiceFnSingleton[I, S] {
	return &ServiceFnSingleton[I, S]{
		fn: fn,
	}
}

// ProvideFactory registers a factory service with pal. See Provide for info on type arguments.
// A new factory service instances are created every time the service is invoked.
// it's the caller's responsibility to shut down the service, pal will also not healthcheck it.
func ProvideFactory[I any, S any]() *ServiceFactory[I, S] {
	return &ServiceFactory[I, S]{}
}

// ProvideFnFactory registers a factory service that is build with a given function.
func ProvideFnFactory[I any]() *ServiceFnFactory[I] {
	return &ServiceFnFactory[I]{}
}

// ProvideConst registers a const as a service.
func ProvideConst[I any](value I) *ServiceConst[I] {
	return &ServiceConst[I]{value}
}

// Invoke retrieves or creates an instance of type I from the given Pal container.
func Invoke[I any](ctx context.Context, invoker Invoker) (I, error) {
	name := elem[I]().String()

	a, err := invoker.Invoke(ctx, name)
	if err != nil {
		return empty[I](), err
	}

	casted, ok := a.(I)
	if !ok {
		return empty[I](), fmt.Errorf("%w: %s. %+v does not implement %s", ErrServiceInvalid, name, a, name)
	}

	return casted, nil
}

// Inject resolves dependencies for a struct of type S using the provided context and Pal instance.
// It initializes the struct's fields by injecting appropriate dependencies based on the field types.
// Returns the fully initialized struct or an error if dependency resolution fails.
func Inject[S any](ctx context.Context, invoker Invoker) (*S, error) {
	s := new(S)
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)

		if !field.CanSet() {
			continue
		}

		fieldType := t.Field(i).Type
		dependency, err := invoker.Invoke(ctx, fieldType.String())
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return nil, err
		}

		field.Set(reflect.ValueOf(dependency))
	}

	return s, nil
}

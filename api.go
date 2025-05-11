package pal

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// Provide registers a singleton service with pal. *I* must be an interface, and *S* must be a struct that implements I.
// Only one instance of the service will be created and reused.
func Provide[I any, S any]() *Service[I, S] {
	return &Service[I, S]{}
}

// ProvideFactory registers a factory service with pal. *I* must be an interface, and *S* must be a struct that implements I.
// A new factory service instances are created every time the service is invoked.
// it's the caller's responsibility to shut down the service, pal will also not healthcheck it.
func ProvideFactory[I any, S any]() *FactoryService[I, S] {
	return &FactoryService[I, S]{}
}

// ProvideConst registers a const as a service.
func ProvideConst[I any](value I) *ConstService[I] {
	return &ConstService[I]{value}
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

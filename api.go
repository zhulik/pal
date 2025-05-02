package pal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/zhulik/pal/pkg/core"
)

// Provide registers a singleton service with pal. *I* must be an interface, and *S* must be a struct that implements I.
// Only one instance of the service will be created and reused.
func Provide[I any, S any]() *Service[I, S] {
	_, isRunner := any(new(S)).(core.Runner)

	return &Service[I, S]{
		singleton: true,
		runner:    isRunner,
	}
}

// ProvideFactory registers a factory service with pal. *I* must be an interface, and *S* must be a struct that implements I.
// A new factory service instances are created every time the service is invoked.
// it's the caller's responsibility to shut down the service, pal will also not healthcheck it.
func ProvideFactory[I any, S any]() *Service[I, S] {
	return &Service[I, S]{
		singleton: false,
	}
}

// Invoke retrieves or creates an instance of type I from the given Pal container.
func Invoke[I any](ctx context.Context, p *Pal) (I, error) {
	name := elem[I]().String()

	a, err := p.Invoke(ctx, name)
	if err != nil {
		return empty[I](), err
	}

	casted, ok := a.(I)
	if !ok {
		return empty[I](), fmt.Errorf("%w: %s. %+v does not implement %s", core.ErrServiceInvalid, a, name, name)
	}

	return casted, nil
}

// Inject resolves dependencies for a struct of type S using the provided context and Pal instance.
// It initializes the struct's fields by injecting appropriate dependencies based on the field types.
// Returns the fully initialized struct or an error if dependency resolution fails.
func Inject[S any](ctx context.Context, p *Pal) (*S, error) {
	s := new(S)
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)

		if !field.CanSet() {
			continue
		}

		fieldType := t.Field(i).Type
		if fieldType.Kind() != reflect.Interface {
			continue
		}

		dependency, err := p.Invoke(ctx, fieldType.String())
		if err != nil {
			return nil, err
		}

		field.Set(reflect.ValueOf(dependency))
	}

	return s, nil
}

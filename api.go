package pal

import (
	"context"
	"fmt"
	"reflect"
)

// New creates and returns a new instance of Pal with the provided ServiceFactory's
func New(factories ...ServiceFactory) *Pal {
	index := make(map[string]ServiceFactory)

	for _, factory := range factories {
		index[factory.Name()] = factory
	}

	return &Pal{
		config:   &Config{},
		store:    newStore(index),
		stopChan: make(chan error),
		log:      func(string, ...any) {},
	}
}

// FromContext retrieves a *Pal from the provided context, expecting it to be stored under the CtxValue key.
// Panics if ctx misses the value.
func FromContext(ctx context.Context) *Pal {
	return ctx.Value(CtxValue).(*Pal)
}

// Provide registers a singleton service with pal. *I* must be an interface and *S* must be a struct that implements I.
// Only one instance of the service will be created and reused.
// TODO: any ways to enforce this with types?
func Provide[I any, S any]() ServiceFactory {
	_, isRunner := any(empty[S]()).(Runner)

	return &serviceFactory[I, S]{
		singleton: true,
		runner:    isRunner,
	}
}

// ProvideFactory registers a factory service with pal. *I* must be an interface, and *S* must be a struct that implements I.
// A new factory service instances are created every time the service is invoked.
// it's the caller's responsibility to shut down the service, pal will also not healthcheck it.
func ProvideFactory[I any, S any]() ServiceFactory {
	return &serviceFactory[I, S]{
		singleton: false,
	}
}

// Invoke retrieves or creates an instance of type I from the given Pal container.
func Invoke[I any](ctx context.Context, p *Pal) (I, error) {
	name := reflect.TypeOf((*I)(nil)).Elem().String()

	a, err := p.Invoke(ctx, name)
	if err != nil {
		return empty[I](), err
	}

	casted, ok := a.(I)
	if !ok {
		return empty[I](), fmt.Errorf("%w: %s. %+v does not implement %s", ErrServiceCastingFailed, a, name, name)
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

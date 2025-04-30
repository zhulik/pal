package pal

import (
	"context"
	"log"
	"reflect"
)

// Provide registers a singleton service with pal. *I* must be an interface and *S* must be a struct that implements I.
// Only one instance of the service will be created and reused.
// TODO: any ways to enforce this with types?
func Provide[I any, S any]() ServiceFactory {
	var s S
	_, isRunner := any(s).(Runner)

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
func Invoke[I any](ctx context.Context, p *Pal) I {
	return p.Invoke(ctx, reflect.TypeOf((*I)(nil)).Elem().String()).(I)
}

func Inject[S any](p *Pal) *S {
	s := new(S)
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)

		if !field.CanSet() {
			log.Printf("%T", s)
			continue
		}

		fieldType := t.Field(i).Type
		if fieldType.Kind() != reflect.Interface {
			continue
		}

		dependency := p.Invoke(context.Background(), fieldType.String())

		field.Set(reflect.ValueOf(dependency))
	}

	return s
}

package pal

import (
	"context"
	"fmt"
)

type serviceFactory[I any, S any] struct {
	singleton bool
}

// String returns a string representation of the dependency.
func (d serviceFactory[I, S]) String() string {
	var i I
	var s S
	return fmt.Sprintf("serviceFactory[%T, %T]", i, s)
}

// Name returns a name of the dependency derived from the interface.
func (d serviceFactory[I, S]) Name() string {
	var i I
	return fmt.Sprintf("%T", i)
}

// Create creates a new instance of the service, if the service implements Runner, it will be run in a background goroutine.
func (d serviceFactory[I, S]) Create(ctx context.Context, p *Pal) (any, error) {
	var s S

	ctx = context.WithValue(ctx, CtxValue, p)

	ctx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	defer cancel()

	if initer, ok := any(s).(Initer); ok {
		if err := initer.Init(ctx); err != nil {
			var i I
			return i, err
		}
	}

	if runner, ok := any(s).(Runner); ok {
		go func() {
			// TODO: use a custom context struct?
			ctx := context.WithValue(context.Background(), CtxValue, p)
			err := runner.Run(ctx)

			if err != nil {
				p.Error(err)
			}
		}()
	}

	return &s, nil
}

func (d serviceFactory[I, S]) IsSingleton() bool {
	return d.singleton
}

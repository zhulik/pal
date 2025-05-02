package main

import (
	"context"
	"time"

	"log"
)

type LeafService interface {
	Bar() string
}

type leafService struct{}

func (s leafService) Bar() string {
	return "bar"
}

func (s leafService) Shutdown(_ context.Context) error {
	return nil
}

type TransientService interface {
	Baz() string
}

type transientService struct {
	Leaf LeafService
}

func (s transientService) Baz() string {
	return s.Leaf.Bar() + "baz"
}

func (s transientService) Shutdown(_ context.Context) error {
	return nil
}

type Service interface {
	Foo() string
}

type service struct {
	Leaf      LeafService
	Transient TransientService

	foo string
}

func (s service) Foo() string {
	return s.Leaf.Bar() + s.Transient.Baz() + s.foo
}

func (s *service) Init(_ context.Context) error { //nolint:unparam
	s.foo = "foo"
	return nil
	// return errors.New("init error")
}

func (s service) Shutdown(_ context.Context) error {
	return nil
}

func (s service) Run(_ context.Context) error { //nolint:unparam
	log.Printf("service: %s", s.Foo())

	time.Sleep(1 * time.Second)

	return nil
}

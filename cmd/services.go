package main

import (
	"context"
	"log"
)

type LeafService interface {
	Bar() string
}

type leafService struct{}

func (s leafService) Bar() string {
	return "bar"
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

func (s *service) Init(_ context.Context) error {
	s.foo = "foo"
	return nil
}

func (s service) Shutdown(_ context.Context) error {
	return nil
}

func (s service) Run(_ context.Context) error {
	log.Println(s.Foo())

	return nil
}

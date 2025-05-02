package pal_test

import (
	"context"
)

// TestInterface is a simple interface for testing
type TestInterface interface {
	DoSomething() string
}

// TestStruct implements TestInterface
type TestStruct struct {
	Value string
}

func (t TestStruct) DoSomething() string {
	return t.Value
}

// RunnerInterface extends TestInterface and core.Runner for testing runner services
type RunnerInterface interface {
	DoSomething() string
}

// RunnerStruct implements RunnerInterface
type RunnerStruct struct {
	RunCalled bool
}

func (r *RunnerStruct) Run(_ context.Context) error {
	r.RunCalled = true
	return nil
}

func (r *RunnerStruct) DoSomething() string {
	return "Something"
}

// DependentStruct is a struct with a dependency on TestInterface
type DependentStruct struct {
	Dependency TestInterface
}

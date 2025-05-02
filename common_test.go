package pal_test

import (
	"context"

	"github.com/zhulik/pal/pkg/core"
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
	TestInterface
	core.Runner
}

// RunnerStruct implements RunnerInterface
type RunnerStruct struct {
	TestStruct
	RunCalled bool
}

func (r *RunnerStruct) Run(_ context.Context) error {
	r.RunCalled = true
	return nil
}

// DependentStruct is a struct with a dependency on TestInterface
type DependentStruct struct {
	Dependency TestInterface
}

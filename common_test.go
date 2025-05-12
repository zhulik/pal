package pal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/zhulik/pal"
)

var (
	errTest = errors.New("test error")
)

// TestServiceInterface is a simple interface for testing
type TestServiceInterface interface {
	DoSomething() string
}

// TestServiceStruct implements TestServiceInterface
type TestServiceStruct struct {
	mock.Mock
}

func (t *TestServiceStruct) HealthCheck(ctx context.Context) error {
	args := t.Called(ctx)
	return args.Error(0)
}
func (t *TestServiceStruct) Init(ctx context.Context) error {
	args := t.Called(ctx)
	return args.Error(0)
}

func (t *TestServiceStruct) Shutdown(ctx context.Context) error {
	args := t.Called(ctx)
	return args.Error(0)
}

func (t *TestServiceStruct) DoSomething() string {
	args := t.Called()
	return args.String(0)
}

// RunnerServiceInterface extends TestServiceInterface and core.Runner for testing runner services
type RunnerServiceInterface interface {
	DoSomething() string
}

// RunnerServiceStruct implements RunnerServiceInterface
type RunnerServiceStruct struct {
	mock.Mock
}

func (r *RunnerServiceStruct) Run(ctx context.Context) error {
	args := r.Called(ctx)
	return args.Error(0)
}

func (r *RunnerServiceStruct) DoSomething() string {
	args := r.Called()
	return args.String(0)
}

// DependentStruct is a struct with a dependency on TestServiceInterface
type DependentStruct struct {
	Dependency TestServiceInterface
}

func eventuallyAssertExpectations(t *testing.T, instance any) {
	t.Helper()

	m := instance.(interface{ AssertExpectations(t mock.TestingT) bool })
	m.AssertExpectations(t)
}

func newPal(services ...pal.ServiceDef) *pal.Pal {
	return pal.New(services...).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)
}

// MockInvoker is a mock implementation of the Invoker interface
type MockInvoker struct {
	mock.Mock
}

func (m *MockInvoker) Invoke(ctx context.Context, name string) (any, error) {
	args := m.Called(ctx, name)
	return args.Get(0), args.Error(1)
}

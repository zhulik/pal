package pal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zhulik/pal"
)

var (
	errTest  = errors.New("test error")
	errTest2 = errors.New("test error 2")
)

// TestServiceInterface is a simple interface for testing
type TestServiceInterface interface {
	DoSomething() string
}

// RunnerServiceInterface is a simple interface for testing
type RunnerServiceInterface interface{}

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

// RunnerServiceStruct is a test helper struct
type RunnerServiceStruct struct {
	*MockRunner
	*MockRunConfiger
}

func NewMockRunnerServiceStruct(t *testing.T) *RunnerServiceStruct {
	return &RunnerServiceStruct{
		MockRunner:      NewMockRunner(t),
		MockRunConfiger: NewMockRunConfiger(t),
	}
}

// DependentStruct is a struct with a dependency on TestServiceInterface
type DependentStruct struct {
	Dependency *TestServiceStruct
}

func eventuallyAssertExpectations(t *testing.T, instance any) {
	t.Helper()

	m := instance.(interface{ AssertExpectations(t mock.TestingT) bool })
	assert.True(t, m.AssertExpectations(t))
}

func newPal(services ...pal.ServiceDef) *pal.Pal {
	return pal.New(services...).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(3 * time.Second)
}

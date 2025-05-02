package pal_test

import (
	"context"
	"errors"
	"sync"
)

// TestStateTracker is used to track the state of service methods during tests
type TestStateTracker struct {
	mu                    sync.Mutex
	shutdownTrackerCalled bool
	errorRunnerCalled     bool
	runnerCalled          bool
}

// NewTestStateTracker creates a new TestStateTracker
func NewTestStateTracker() *TestStateTracker {
	return &TestStateTracker{}
}

// SetShutdownTrackerCalled sets the shutdownTrackerCalled flag
func (t *TestStateTracker) SetShutdownTrackerCalled() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.shutdownTrackerCalled = true
}

// ShutdownTrackerCalled returns the shutdownTrackerCalled flag
func (t *TestStateTracker) ShutdownTrackerCalled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.shutdownTrackerCalled
}

// SetErrorRunnerCalled sets the errorRunnerCalled flag
func (t *TestStateTracker) SetErrorRunnerCalled() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorRunnerCalled = true
}

// ErrorRunnerCalled returns the errorRunnerCalled flag
func (t *TestStateTracker) ErrorRunnerCalled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.errorRunnerCalled
}

// SetRunnerCalled sets the runnerCalled flag
func (t *TestStateTracker) SetRunnerCalled() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.runnerCalled = true
}

// RunnerCalled returns the runnerCalled flag
func (t *TestStateTracker) RunnerCalled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.runnerCalled
}

// testStateKey is the key used to store the TestStateTracker in the context
type testStateKey struct{}

// WithTestState adds a TestStateTracker to the context
func WithTestState(ctx context.Context, tracker *TestStateTracker) context.Context {
	return context.WithValue(ctx, testStateKey{}, tracker)
}

// GetTestState retrieves the TestStateTracker from the context
func GetTestState(ctx context.Context) *TestStateTracker {
	if tracker, ok := ctx.Value(testStateKey{}).(*TestStateTracker); ok {
		return tracker
	}
	return nil
}

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

func (r *RunnerStruct) Run(ctx context.Context) error {
	r.RunCalled = true

	if tracker := GetTestState(ctx); tracker != nil {
		tracker.SetRunnerCalled()
	}

	return nil
}

func (r *RunnerStruct) DoSomething() string {
	return "Something"
}

// DependentStruct is a struct with a dependency on TestInterface
type DependentStruct struct {
	Dependency TestInterface
}

// FailingInitInterface is an interface for a service that fails during initialization
type FailingInitInterface interface {
	DoSomething() string
}

// FailingInitStruct implements FailingInitInterface and fails during initialization
type FailingInitStruct struct {
	InitCalled bool
	Dependency ShutdownTrackingInterface
}

func (f *FailingInitStruct) Init(_ context.Context) error {
	f.InitCalled = true
	return errors.New("init error")
}

func (f *FailingInitStruct) DoSomething() string {
	return "Something"
}

// ShutdownTrackingInterface is an interface for a service that tracks if it was shut down
type ShutdownTrackingInterface interface {
	DoSomething() string
	WasShutDown() bool
}

// ShutdownTrackingStruct implements ShutdownTrackingInterface
type ShutdownTrackingStruct struct {
	ShutdownCalled bool
}

func (s *ShutdownTrackingStruct) Init(_ context.Context) error {
	return nil
}

func (s *ShutdownTrackingStruct) Shutdown(ctx context.Context) error {
	s.ShutdownCalled = true

	if tracker := GetTestState(ctx); tracker != nil {
		tracker.SetShutdownTrackerCalled()
	}

	return nil
}

func (s *ShutdownTrackingStruct) DoSomething() string {
	return "Something"
}

func (s *ShutdownTrackingStruct) WasShutDown() bool {
	return s.ShutdownCalled
}

// ErrorRunnerInterface is an interface for a runner that returns an error
type ErrorRunnerInterface interface {
	DoSomething() string
}

// ErrorRunnerStruct implements ErrorRunnerInterface and returns an error from Run
type ErrorRunnerStruct struct {
	RunCalled bool
}

func (r *ErrorRunnerStruct) Run(ctx context.Context) error {
	r.RunCalled = true

	if tracker := GetTestState(ctx); tracker != nil {
		tracker.SetErrorRunnerCalled()
	}

	return errors.New("run error")
}

func (r *ErrorRunnerStruct) DoSomething() string {
	return "Something"
}

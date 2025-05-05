package pal_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zhulik/pal"
)

// SimpleDataService is a simple service that provides data
type SimpleDataService interface {
	GetData() string
}

// SimpleDataServiceImpl implements SimpleDataService
type SimpleDataServiceImpl struct{}

// GetData returns some data
func (s *SimpleDataServiceImpl) GetData() string {
	return "simple data"
}

// This example demonstrates how to use Pal with a simple service.
func Example_pal_simple() {
	// Create a Pal instance with a simple service
	p := pal.New(
		pal.Provide[SimpleDataService, SimpleDataServiceImpl](),
	)

	// Initialize Pal
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		fmt.Printf("Failed to initialize Pal: %v\n", err)
		return
	}

	// Invoke the service
	instance, err := p.Invoke(ctx, "pal_test.SimpleDataService")
	if err != nil {
		fmt.Printf("Failed to invoke service: %v\n", err)
		return
	}

	// Use the service
	service := instance.(SimpleDataService)
	fmt.Println(service.GetData())

	// Output: simple data
}

// InitService is a service that implements the Initer interface
type InitService interface {
	GetStatus() string
}

// InitServiceImpl implements InitService and pal.Initer
type InitServiceImpl struct {
	initialized bool
}

// Init initializes the service
func (s *InitServiceImpl) Init(_ context.Context) error {
	s.initialized = true
	return nil
}

// GetStatus returns the initialization status
func (s *InitServiceImpl) GetStatus() string {
	if s.initialized {
		return "initialized"
	}
	return "not initialized"
}

// This example demonstrates how to use Pal with a service that implements the Initer interface.
func Example_pal_initer() {
	// Create a Pal instance with a service that implements Initer
	p := pal.New(
		pal.Provide[InitService, InitServiceImpl](),
	)

	// Initialize Pal
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		fmt.Printf("Failed to initialize Pal: %v\n", err)
		return
	}

	// Invoke the service
	instance, err := p.Invoke(ctx, "pal_test.InitService")
	if err != nil {
		fmt.Printf("Failed to invoke service: %v\n", err)
		return
	}

	// Use the service
	service := instance.(InitService)
	fmt.Println(service.GetStatus())

	// Output: initialized
}

// BackgroundTask is a service that runs in the background
type BackgroundTask interface {
	GetStatus() string
}

// BackgroundTaskImpl implements BackgroundTask and pal.Runner
type BackgroundTaskImpl struct {
	status string
	mu     sync.Mutex
}

// Run runs the background task
func (s *BackgroundTaskImpl) Run(_ context.Context) error {
	s.mu.Lock()
	s.status = "running"
	s.mu.Unlock()
	// In a real application, this would do some work
	return nil
}

// GetStatus returns the status of the task
func (s *BackgroundTaskImpl) GetStatus() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

// This example demonstrates how to use Pal with a runner service.
func Example_pal_runner() {
	// Create a Pal instance with a runner service
	p := pal.New(
		pal.Provide[BackgroundTask, BackgroundTaskImpl](),
	)

	// Set timeouts
	p.InitTimeout(1 * time.Second)
	p.ShutdownTimeout(1 * time.Second)

	// Create a context that will be canceled after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Initialize Pal
	if err := p.Init(ctx); err != nil {
		fmt.Printf("Failed to initialize Pal: %v\n", err)
		return
	}

	// Invoke the service to check its status
	instance, err := p.Invoke(ctx, "pal_test.BackgroundTask")
	if err != nil {
		fmt.Printf("Failed to invoke service: %v\n", err)
		return
	}

	// Use the service
	service := instance.(BackgroundTask)

	// Start the runner
	go func() {
		err := p.Run(ctx)
		if err != nil {
			fmt.Printf("Pal.Run returned error: %v\n", err)
		}
	}()

	// Give the runner a moment to start
	time.Sleep(50 * time.Millisecond)

	// Check the status
	fmt.Println(service.GetStatus())

	// Output: running
}

package pal_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zhulik/pal"
)

// SimpleService is a test service interface
type SimpleService interface {
	GetMessage() string
}

// SimpleServiceImpl implements SimpleService
type SimpleServiceImpl struct{}

// GetMessage returns a greeting message
func (s *SimpleServiceImpl) GetMessage() string {
	return "Hello from SimpleService"
}

// This example demonstrates how to create a Pal instance with services and use it.
func Example_container() {
	// Create a Pal instance with the service
	p := pal.New(
		pal.Provide[SimpleService](&SimpleServiceImpl{}),
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)

	// Initialize Pal
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		fmt.Printf("Failed to initialize Pal: %v\n", err)
		return
	}

	// Invoke the service
	instance, err := p.Invoke(ctx, "pal_test.SimpleService")
	if err != nil {
		fmt.Printf("Failed to invoke service: %v\n", err)
		return
	}

	// Use the service
	service := instance.(SimpleService)
	fmt.Println(service.GetMessage())

	// Output: Hello from SimpleService
}

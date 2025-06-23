package pal_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zhulik/pal"
)

// ExampleService is a service that runs in the background
type ExampleService interface {
	// This is a public API exposed to other components of the app.
	// No need to include pal interfaces here.
	Foo() string
}

// ExampleServiceImpl implements ExampleService and pal.Runner
type ExampleServiceImpl struct {
}

// Init initializes the services
func (s *ExampleServiceImpl) Init(_ context.Context) error {
	fmt.Println("init")

	// In a real app, here you'd create resources or open connections to other services and databases.

	return nil
}

// Run runs the background task
func (s *ExampleServiceImpl) Run(_ context.Context) error {
	fmt.Println("run")

	// In a real application, this would do some work
	return nil
}

// Shutdown initializes the services
func (s *ExampleServiceImpl) Shutdown(_ context.Context) error {
	fmt.Println("shutdown")

	// In a real app, here you'd release resources or close connections to other services and databases.

	return nil
}

// Foo does foo.
func (s *ExampleServiceImpl) Foo() string {
	// In case of a regular service, you put your actual application logic here.
	// In case of a Runner - you may want to interact with the background job via channels, this is the place.
	return "foo"
}

// This example demonstrates how to use Pal with a runner service.
func Example_pal_runner() {
	p := pal.New(
		pal.Provide[ExampleService](&ExampleServiceImpl{}),
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)

	err := p.Run(context.Background())
	if err != nil {
		fmt.Printf("Pal.Run returned error: %v\n", err)
	}

	// Output:
	// init
	// run
	// shutdown
}

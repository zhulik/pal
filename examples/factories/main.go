package main

import (
	"context"
	"time"

	"github.com/zhulik/pal"
)

// Pinger is an interface for the pinger service.
type Pinger interface {
	Ping(ctx context.Context) error
}

// main is the entry point of the program.
func main() {
	// Create a new pal application, provide the services and initialize pal's lifecycle timeouts.
	// Pal is aware of the dependencies between the services and initlizes them in correct order:
	// first pinger, then ticker. After initialization, it runs the runners, in this case the ticker
	// When shutting down, it first shuts down the ticker, then the pinger. First it stops the runners,
	// then shuts down the services in the order reversed to the initialization.
	p := pal.New(
		pal.ProvideFactory1(func(_ context.Context, url string) (Pinger, error) {
			return &pinger{URL: url}, nil
		}), // Provide the pinger service as the Pinger interface.
		pal.Provide(&ticker{}), // Provide the ticker service. As it is the main runner, it does not need to have a specific interface.
	).
		InjectSlog().                    // Enables automatic logger injection.
		InitTimeout(time.Second).        // Set the timeout for the initialization phase.
		HealthCheckTimeout(time.Second). // Set the timeout for the health check phase.
		ShutdownTimeout(3 * time.Second) // Set the timeout for the shutdown phase.

	if err := p.Run(context.Background()); err != nil { // Run the application.
		panic(err)
	}
}

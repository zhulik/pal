package main

import (
	"context"
	"time"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/examples/web/pinger"
	"github.com/zhulik/pal/examples/web/server"
)

// main is the entry point of the program.
func main() {
	// Create a new pal application, provide the services and initialize pal's lifecycle timeouts.
	// Pal is aware of the dependencies between the services and initlizes them in correct order:
	// first pinger, then server. After initialization, it runs the runners, in this case the server
	// When shutting down, it first shuts down the server, then the pinger. First it stops the runners,
	// then shuts down the services in the order reversed to the initialization.
	p := pal.New(
		pinger.Provide(), // Provide services from the pinger module.
		server.Provide(), // Provide services from the server module.
	).
		InjectSlog().                              // Enables automatic logger injection.
		RunHealthCheckServer(":8081", "/healthz"). // Run the health check server.
		InitTimeout(time.Second).                  // Set the timeout for the initialization phase.
		HealthCheckTimeout(time.Second).           // Set the timeout for the health check phase.
		ShutdownTimeout(3 * time.Second)           // Set the timeout for the shutdown phase.

	if err := p.Run(context.Background()); err != nil { // Run the application.
		panic(err)
	}
}

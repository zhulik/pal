package main

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/inspect"
)

type Inspect interface {
}
type Console interface{}

// This example demonstrates how to use Pal with a runner service.
func main() {
	p := pal.New(
		pal.Provide[Inspect, inspect.Inspect](),
		pal.Provide[Console, inspect.Console](),
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)

	err := p.Run(context.Background(), syscall.SIGINT)
	if err != nil {
		fmt.Printf("Pal.Run returned error: %v\n", err)
	}

	// Output:
	// init
	// run
	// shutdown
}

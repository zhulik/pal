package inspect_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/inspect"
)

type Inspect interface {
}

// This example demonstrates how to use Pal with a runner service.
func Example_inspect() {
	p := pal.New(
		pal.Provide[Inspect, inspect.Inspect](),
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

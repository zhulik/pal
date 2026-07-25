package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zhulik/pal"
)

func main() {
	p := pal.New().
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(3 * time.Second)

	err := p.Run(context.Background())
	if err != nil {
		fmt.Printf("Pal.Run returned error: %v\n", err)
	}
}

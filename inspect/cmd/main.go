package main

import (
	"context"
	"log"
	"syscall"
	"time"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/inspect"
)

func main() {
	p := pal.New(
		inspect.ProvideBase(),
		inspect.ProvideRemoteConsole(),
	).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)

	err := p.Run(context.Background(), syscall.SIGINT)
	if err != nil {
		log.Fatalf("Pal.Run returned error: %v\n", err)
	}
}

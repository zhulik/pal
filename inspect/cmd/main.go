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
	services := append(inspect.Provide(),
		pal.Provide[*inspect.RemoteConsole, inspect.RemoteConsole](),
	)
	p := pal.New(services...).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(time.Second)

	err := p.Run(context.Background(), syscall.SIGINT)
	if err != nil {
		log.Fatalf("Pal.Run returned error: %v\n", err)
	}
}

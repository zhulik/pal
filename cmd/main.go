package main

import (
	"context"
	"syscall"
	"time"

	"github.com/zhulik/pal"
	"log"
)

func main() {
	err := pal.New(
		pal.Provide[Service, service](),
		pal.Provide[LeafService, leafService](),
		pal.Provide[TransientService, transientService](),
		// pal.ProvideFactory[Service, *service](),
	).
		SetLogger(log.Printf).
		InitTimeout(3*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(3*time.Second).
		Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		log.Fatal(err)
	}
}

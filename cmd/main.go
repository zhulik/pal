package main

import (
	"context"
	"syscall"
	"time"

	"log"

	"github.com/zhulik/pal"
)

func main() {
	err := pal.New(
		pal.Provide[Service, service]().BeforeInit(func(_ context.Context, _ *service) error {
			log.Printf("service before init")
			return nil
		}),
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

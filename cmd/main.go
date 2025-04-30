package main

import (
	"context"
	"errors"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/zhulik/pal"
)

func main() {
	err := pal.New(
		pal.Provide[Service, service](),
		pal.Provide[LeafService, leafService](),
		pal.Provide[TransientService, transientService](),
		// pal.ProvideFactory[Service, *service](),
	).
		InitTimeout(3*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(3*time.Second).
		Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		var palError *pal.RunError
		errors.As(err, &palError)
		log.Println(err)
		os.Exit(palError.ExitCode())
	}
}

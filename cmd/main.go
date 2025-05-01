package main

import (
	"context"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zhulik/pal"
)

func main() {
	err := pal.New(
		pal.Provide[Service, service](),
		pal.Provide[LeafService, leafService](),
		pal.Provide[TransientService, transientService](),
		// pal.ProvideFactory[Service, *service](),
	).
		SetLogger(log.WithField("component", "pal").Infof).
		InitTimeout(3*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(3*time.Second).
		Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		log.Fatal(err)
	}
}

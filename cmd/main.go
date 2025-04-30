package main

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"

	"github.com/zhulik/pal"
)

type Service interface {
	Foo() string
}

type service struct {
}

func (s *service) Foo() string {
	return "bar"
}

func (s *service) HealthCheck(_ context.Context) error {
	return nil
}

func (s *service) Shutdown(_ context.Context) error {
	return nil
}

func (s *service) Run(_ context.Context) error {
	return nil
}

func (s *service) Init(_ context.Context) error {
	return nil
}

func main() {
	err := pal.New(
		pal.Provide[Service, *service](),
		pal.ProvideFactory[Service, *service](),
	).
		InitTimeout(3*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(3*time.Second).
		Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		var palError *pal.RunError
		errors.As(err, &palError)
		os.Exit(palError.ExitCode())
	}
}

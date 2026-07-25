package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/zhulik/pal"
)

type runner struct {
	Logger *slog.Logger
}

func (r *runner) Run(ctx context.Context) error {
	r.Logger.Info("{{.Package}} running")
	return nil
}

func main() {
	p := pal.New(
		pal.Provide(&runner{}),
	).
		InjectSlog().
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(3 * time.Second)
	err := p.Run(context.Background())
	if err != nil {
		fmt.Printf("{{.Package}} failed: %v\n", err)
		os.Exit(1)
	}
}

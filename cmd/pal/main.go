package main

import (
	"context"
	"fmt"
	"os"

	"github.com/zhulik/pal/cmd/pal/internal/cli"
)

func main() {
	if err := cli.New().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

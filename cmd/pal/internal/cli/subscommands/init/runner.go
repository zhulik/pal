package initcmd

import (
	"context"
	"fmt"
)

type runner struct {
}

func (r *runner) Run(ctx context.Context) error {
	fmt.Println("init")
	return nil
}

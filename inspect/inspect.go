package inspect

import (
	"context"
)

type Inspect struct{}

func (i *Inspect) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

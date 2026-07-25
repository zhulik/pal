package initcmd

import (
	"context"
	"fmt"
	"os"
)

type runner struct {
	opts Options
}

func (r *runner) Run(ctx context.Context) error {
	_ = ctx

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	opts := r.opts
	if !opts.NoInteractive {
		opts, err = RunWizard(cwd)
		if err != nil {
			return err
		}
	}

	empty, err := IsEmpty(cwd)
	if err != nil {
		return err
	}
	if !empty && !opts.Force {
		return ErrNotEmpty
	}

	if opts.Git {
		if err := initGit(ctx, cwd); err != nil {
			return err
		}
	}

	fmt.Println("Project initialized.")
	return nil
}

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
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	opts := r.opts
	if !opts.NoInteractive {
		// Preserve Module from the CLI argument so the wizard can skip that step.
		opts, err = RunWizard(cwd, Options{Module: opts.Module})
		if err != nil {
			return err
		}
	}

	if opts.Module == "" {
		return ErrModuleRequired
	}

	empty, err := IsEmpty(cwd)
	if err != nil {
		return err
	}
	if !empty && !opts.Force {
		return ErrNotEmpty
	}

	if err := initModule(ctx, cwd, opts.Module); err != nil {
		return err
	}

	if opts.Git {
		if err := initGit(ctx, cwd); err != nil {
			return err
		}
	}

	fmt.Println("Project initialized.")
	return nil
}

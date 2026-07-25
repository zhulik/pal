package initcmd

import (
	"context"
	"fmt"
)

type runner struct {
	opts Options
}

func (r *runner) Run(ctx context.Context) error {
	workDir, promoteTo, cleanup, err := resolveWorkDir(r.opts.Directory)
	if err != nil {
		return err
	}
	defer cleanup()

	empty, err := IsEmpty(workDir)
	if err != nil {
		return err
	}
	if !empty {
		return ErrNotEmpty
	}

	opts := r.opts
	if !opts.NoInteractive {
		// Preserve Module from the CLI argument so the wizard can skip that step.
		opts, err = RunWizard(workDir, Options{Module: opts.Module})
		if err != nil {
			return err
		}
	}

	if opts.Module == "" {
		return ErrModuleRequired
	}

	if err := initModule(ctx, workDir, opts.Module); err != nil {
		return err
	}

	if opts.Git {
		if err := initGit(ctx, workDir); err != nil {
			return err
		}
	}

	if err := applyTemplate(workDir, opts); err != nil {
		return err
	}

	if err := tidyModule(ctx, workDir); err != nil {
		return err
	}

	if promoteTo != "" {
		if err := promoteWorkDir(workDir, promoteTo); err != nil {
			return err
		}
	}

	fmt.Println("Project initialized.")
	return nil
}

package initcmd

import (
	"context"
	"fmt"
)

type runner struct {
	opts Options
}

// Run initializes a project with opts. Prefer this from tests and other
// callers; the CLI Action goes through app.Run → pal with an unexported runner.
func Run(ctx context.Context, opts Options) error {
	return (&runner{opts: opts}).Run(ctx)
}

func (r *runner) Run(ctx context.Context) error {
	workDir, promoteTo, cleanup, err := ResolveWorkDir(r.opts.Directory)
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
		// Seed the wizard from CLI options so explicitly set flags/args skip
		// their steps (via Applicable) while unset ones are still prompted.
		opts, err = RunWizard(workDir, opts)
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
		if err := PromoteWorkDir(workDir, promoteTo); err != nil {
			return err
		}
	}

	fmt.Println("Project initialized.")
	return nil
}

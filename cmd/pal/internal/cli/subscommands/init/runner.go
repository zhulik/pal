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
	workDir, promoteTo, targetExisted, cleanup, err := ResolveWorkDir(r.opts.Directory)
	if err != nil {
		return err
	}
	defer cleanup()

	if targetExisted {
		empty, err := IsEmpty(promoteTo)
		if err != nil {
			return err
		}
		if !empty {
			return ErrNotEmpty
		}
	}

	opts := r.opts
	if !opts.NoInteractive {
		// Seed the wizard from CLI options so explicitly set flags/args skip
		// their steps (via Applicable) while unset ones are still prompted.
		// Pass the final target path (not the staging dir) as wizard context.
		opts, err = RunWizard(promoteTo, opts)
		if err != nil {
			return err
		}
	}
	opts.Description = descriptionOrDefault(opts.Description)

	if opts.Module == "" {
		return ErrModuleRequired
	}
	if err := preflight(opts); err != nil {
		return err
	}

	// All mutations happen only in the staging directory. The target is
	// updated solely via PromoteWorkDir on success.
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

	if err := PromoteWorkDir(workDir, promoteTo, targetExisted); err != nil {
		return err
	}

	// Lint-fix runs in the final target directory (Taskfile + .tool-versions
	// context), after promote, so tooling resolves against the new project.
	if err := lintFix(ctx, promoteTo); err != nil {
		return err
	}

	fmt.Println("Project initialized.")
	return nil
}

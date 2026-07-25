package initcmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"

	"github.com/zhulik/pal/cmd/pal/pkg/cmdexec"
)

const argModule = "module"

type wizardStepModule struct{}

func (wizardStepModule) Flag() cli.Flag {
	return nil
}

func (wizardStepModule) Argument() cli.Argument {
	return &cli.StringArg{
		Name:      argModule,
		UsageText: "MODULE",
	}
}

func (wizardStepModule) Applicable(opts *Options, _ string) bool {
	// Skip when the module was already provided as a CLI argument.
	return opts.Module == ""
}

func (wizardStepModule) Field(opts *Options) (huh.Field, error) {
	return huh.NewInput().
		Title("What Go module path should the new app use?").
		Placeholder("github.com/user/app").
		Value(&opts.Module).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("module path is required")
			}
			return nil
		}), nil
}

func (wizardStepModule) Abort(_ *Options) bool {
	return false
}

func initModule(ctx context.Context, dir, module string) error {
	if err := cmdexec.Run(ctx, dir, "go", "mod", "init", module); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}
	return nil
}

func tidyModule(ctx context.Context, dir string) error {
	if err := cmdexec.Run(ctx, dir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func lintFix(ctx context.Context, dir string) error {
	// Resolve to an absolute path and run with that as the process working
	// directory so Taskfile, asdf (.tool-versions), and golangci-lint all
	// resolve against the generated project — not the pal CLI process cwd.
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory for lint-fix: %w", err)
	}
	if err := cmdexec.Run(ctx, abs, "task", "-d", abs, "lint-fix"); err != nil {
		return fmt.Errorf("task lint-fix: %w", err)
	}
	return nil
}

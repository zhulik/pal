package initcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

const argModule = "module"

type moduleStep struct{}

func (moduleStep) Flag() cli.Flag {
	return nil
}

func (moduleStep) Argument() cli.Argument {
	return &cli.StringArg{
		Name:      argModule,
		UsageText: "MODULE",
	}
}

func (moduleStep) Applicable(opts *Options, _ string) bool {
	// Skip when the module was already provided as a CLI argument.
	return opts.Module == ""
}

func (moduleStep) Field(opts *Options) huh.Field {
	return huh.NewInput().
		Title("What Go module path should the new app use?").
		Placeholder("github.com/user/app").
		Value(&opts.Module).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("module path is required")
			}
			return nil
		})
}

func (moduleStep) Abort(_ *Options) bool {
	return false
}

func initModule(ctx context.Context, dir, module string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "init", module)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod init: %w: %s", err, bytes.TrimSpace(out))
	}
	return nil
}

func tidyModule(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod tidy: %w: %s", err, bytes.TrimSpace(out))
	}
	return nil
}

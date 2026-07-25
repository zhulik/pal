package initcmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"

	"github.com/zhulik/pal/cmd/pal/pkg/cmdexec"
)

const flagNoGit = "no-git"

type wizardStepGit struct{}

func (wizardStepGit) Flag() cli.Flag {
	return &cli.BoolFlag{
		Name:  flagNoGit,
		Usage: "do not initialize a git repository",
	}
}

func (wizardStepGit) Argument() cli.Argument {
	return nil
}

func (wizardStepGit) Applicable(opts *Options, _ string) bool {
	// Skip when --no-git was explicitly passed on the CLI.
	return !opts.GitSet
}

func (wizardStepGit) Field(opts *Options) (huh.Field, error) {
	opts.Git = true
	return huh.NewConfirm().
		Title("Create a git repository?").
		Affirmative("Y").
		Negative("n").
		Value(&opts.Git), nil
}

func (wizardStepGit) Abort(_ *Options) bool {
	return false
}

func initGit(ctx context.Context, dir string) error {
	if err := cmdexec.Run(ctx, dir, "git", "init"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	return nil
}

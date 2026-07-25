package initcmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

const flagNoGit = "no-git"

type gitStep struct{}

func (gitStep) Flag() cli.Flag {
	return &cli.BoolFlag{
		Name:  flagNoGit,
		Usage: "do not initialize a git repository",
	}
}

func (gitStep) Applicable(_ *Options, _ string) bool {
	return true
}

func (gitStep) Field(opts *Options) huh.Field {
	opts.Git = true
	return huh.NewConfirm().
		Title("Create a git repository?").
		Affirmative("Y").
		Negative("n").
		Value(&opts.Git)
}

func initGit(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init: %w: %s", err, bytes.TrimSpace(out))
	}
	return nil
}

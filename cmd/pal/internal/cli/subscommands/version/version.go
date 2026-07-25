package versioncmd

import (
	"context"

	"github.com/urfave/cli/v3"
)

// New returns the version subcommand. It prints the same output as
// --version / -v and does not run through app.Run (no Pal lifecycle).
func New() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "print the pal CLI version",
		Action: func(_ context.Context, cmd *cli.Command) error {
			cli.ShowVersion(cmd.Root())
			return nil
		},
	}
}

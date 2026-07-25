package initcmd

import (
	"context"

	"github.com/urfave/cli/v3"
)

// New returns the init subcommand.
func New() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "initialize a new pal project",
		Flags:     Flags(),
		Arguments: Arguments(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return (&runner{opts: OptionsFromCommand(cmd)}).Run(ctx)
		},
	}
}

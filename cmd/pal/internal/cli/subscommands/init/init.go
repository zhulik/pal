package initcmd

import (
	"context"

	"github.com/urfave/cli/v3"
	"github.com/zhulik/pal"
	"github.com/zhulik/pal/cmd/pal/internal/cli/app"
)

// New returns the init subcommand.
func New() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "initialize a new pal project",
		Flags: Flags(),
		Action: func(context.Context, *cli.Command) error {
			return app.Run(context.Background(), pal.Provide(&runner{}))
		},
	}
}

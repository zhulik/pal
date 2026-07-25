package cli

import (
	"github.com/urfave/cli/v3"

	initcmd "github.com/zhulik/pal/cmd/pal/internal/cli/subscommands/init"
)

// New returns the root pal CLI command with all subcommands registered.
//
// Help is provided by urfave/cli (`help`, `-h`, `--help`); no custom help
// subcommand is needed.
func New() *cli.Command {
	return &cli.Command{
		Name:  "pal",
		Usage: "pal dependency injection toolkit",
		Commands: []*cli.Command{
			initcmd.New(),
		},
	}
}

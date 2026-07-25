package cli

import (
	"github.com/urfave/cli/v3"

	initcmd "github.com/zhulik/pal/cmd/pal/internal/cli/subscommands/init"
	versioncmd "github.com/zhulik/pal/cmd/pal/internal/cli/subscommands/version"
	"github.com/zhulik/pal/cmd/pal/internal/version"
)

// New returns the root pal CLI command with all subcommands registered.
//
// Help is provided by urfave/cli (`help`, `-h`, `--help`); no custom help
// subcommand is needed. Version is provided by urfave/cli (`-v`, `--version`)
// and the `version` subcommand; both print the same string from package
// version.
func New() *cli.Command {
	return &cli.Command{
		Name:    "pal",
		Usage:   "pal dependency injection toolkit",
		Version: version.String(),
		Commands: []*cli.Command{
			initcmd.New(),
			versioncmd.New(),
		},
	}
}

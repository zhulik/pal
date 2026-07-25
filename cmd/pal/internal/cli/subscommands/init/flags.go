package initcmd

import "github.com/urfave/cli/v3"

// Flags returns flags specific to the init subcommand.
func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "interactive",
			Aliases: []string{"i"},
			Usage:   "run init interactively",
		},
	}
}

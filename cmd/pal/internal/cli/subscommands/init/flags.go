package initcmd

import "github.com/urfave/cli/v3"

const flagNoInteractive = "no-interactive"

// Flags returns flags specific to the init subcommand.
func Flags() []cli.Flag {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  flagNoInteractive,
			Usage: "run without the interactive wizard (read options from flags)",
		},
	}
	for _, step := range Steps() {
		flags = append(flags, step.Flag())
	}
	return flags
}

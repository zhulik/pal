package initcmd

import "github.com/urfave/cli/v3"

const (
	flagNoInteractive = "no-interactive"
	flagDirectory     = "directory"
)

// Flags returns flags specific to the init subcommand.
func Flags() []cli.Flag {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  flagNoInteractive,
			Usage: "run without the interactive wizard (require MODULE; other options from flags)",
		},
		&cli.StringFlag{
			Name:    flagDirectory,
			Aliases: []string{"d"},
			Usage:   "project directory (must be empty if it exists; created on success if missing)",
		},
	}
	for _, step := range Steps() {
		if f := step.Flag(); f != nil {
			flags = append(flags, f)
		}
	}
	return flags
}

// Arguments returns positional arguments owned by init steps.
func Arguments() []cli.Argument {
	var args []cli.Argument
	for _, step := range Steps() {
		if a := step.Argument(); a != nil {
			args = append(args, a)
		}
	}
	return args
}

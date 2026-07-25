package initcmd

import (
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

type forceStep struct{}

func (forceStep) Flag() cli.Flag {
	return &cli.BoolFlag{
		Name:    "force",
		Aliases: []string{"f"},
		Usage:   "allow overwriting existing files",
	}
}

func (forceStep) Applicable(_ *Options, cwd string) bool {
	empty, err := IsEmpty(cwd)
	if err != nil {
		// Surface the prompt when we cannot determine emptiness; the backend
		// will still validate.
		return true
	}
	return !empty
}

func (forceStep) Field(opts *Options) huh.Field {
	return huh.NewConfirm().
		Title("Directory is not empty. Continue and allow overwriting?").
		Affirmative("y").
		Negative("N").
		Value(&opts.Force)
}

package initcmd

import (
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

const (
	flagDescription    = "description"
	defaultDescription = "//TODO: add description"
)

type wizardStepDescription struct{}

func (wizardStepDescription) Flag() cli.Flag {
	return &cli.StringFlag{
		Name:  flagDescription,
		Usage: "short project description for README and docs (default: " + defaultDescription + ")",
	}
}

func (wizardStepDescription) Argument() cli.Argument {
	return nil
}

func (wizardStepDescription) Applicable(opts *Options, _ string) bool {
	return opts.Description == ""
}

func (wizardStepDescription) Field(opts *Options) (huh.Field, error) {
	return huh.NewInput().
		Title("Short project description?").
		Placeholder(defaultDescription).
		Value(&opts.Description), nil
}

func (wizardStepDescription) Abort(_ *Options) bool {
	return false
}

func descriptionOrDefault(s string) string {
	if s == "" {
		return defaultDescription
	}
	return s
}

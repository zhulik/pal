package initcmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
	"github.com/zhulik/pal/cmd/pal/templates"
)

const (
	flagTemplate    = "template"
	defaultTemplate = "cli"
)

type templateStep struct{}

func (templateStep) Flag() cli.Flag {
	return &cli.StringFlag{
		Name:  flagTemplate,
		Usage: "project template to use",
		Value: defaultTemplate,
	}
}

func (templateStep) Argument() cli.Argument {
	return nil
}

func (templateStep) Applicable(_ *Options, _ string) bool {
	return true
}

func (templateStep) Field(opts *Options) huh.Field {
	if opts.Template == "" {
		opts.Template = defaultTemplate
	}
	return huh.NewSelect[string]().
		Title("Which project template?").
		Options(huh.NewOption("CLI", "cli")).
		Value(&opts.Template)
}

func (templateStep) Abort(_ *Options) bool {
	return false
}

func applyTemplate(dir string, opts Options) error {
	name := opts.Template
	if name == "" {
		name = defaultTemplate
	}
	if err := templates.Apply(dir, name, templates.Data{
		Package: templates.PackageName(opts.Module),
	}); err != nil {
		return fmt.Errorf("apply template %q: %w", name, err)
	}
	return nil
}

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

func (templateStep) Applicable(opts *Options, _ string) bool {
	// Skip when --template was explicitly passed on the CLI.
	return !opts.TemplateSet
}

func (templateStep) Field(opts *Options) (huh.Field, error) {
	if opts.Template == "" {
		opts.Template = defaultTemplate
	}
	names, err := templates.Names()
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no project templates embedded")
	}
	options := make([]huh.Option[string], 0, len(names))
	for _, name := range names {
		options = append(options, huh.NewOption(name, name))
	}
	return huh.NewSelect[string]().
		Title("Which project template?").
		Options(options...).
		Value(&opts.Template), nil
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
		Module:  opts.Module,
	}); err != nil {
		return fmt.Errorf("apply template %q: %w", name, err)
	}
	return nil
}

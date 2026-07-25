package initcmd

import (
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

// Step owns both a CLI input (flag or positional argument) and its
// corresponding wizard field so interactive and non-interactive modes stay in
// sync.
type Step interface {
	// Flag returns the CLI flag for this step, or nil if it uses a positional
	// argument instead.
	Flag() cli.Flag
	// Argument returns the positional argument for this step, or nil if it uses
	// a flag instead.
	Argument() cli.Argument
	// Applicable reports whether the wizard should ask this step.
	Applicable(opts *Options, cwd string) bool
	// Field returns a huh field bound to opts, or an error if the step cannot
	// be presented (e.g. template discovery failed).
	Field(opts *Options) (huh.Field, error)
	// Abort reports whether the wizard should stop after this step was answered
	// (fail fast). Only called when the step was actually shown.
	Abort(opts *Options) bool
}

// Steps returns the ordered registry of init steps.
func Steps() []Step {
	return []Step{
		wizardStepModule{},
		wizardStepTemplate{},
		wizardStepGit{},
	}
}

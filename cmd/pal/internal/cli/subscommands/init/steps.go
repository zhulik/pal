package initcmd

import (
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

// Step owns both a CLI flag and its corresponding wizard field so interactive
// and non-interactive modes stay in sync.
type Step interface {
	Flag() cli.Flag
	// Applicable reports whether the wizard should ask this step.
	Applicable(opts *Options, cwd string) bool
	// Field returns a huh field bound to opts.
	Field(opts *Options) huh.Field
	// Abort reports whether the wizard should stop after this step was answered
	// (fail fast). Only called when the step was actually shown.
	Abort(opts *Options) bool
}

// Steps returns the ordered registry of init steps.
func Steps() []Step {
	return []Step{
		forceStep{},
		gitStep{},
	}
}

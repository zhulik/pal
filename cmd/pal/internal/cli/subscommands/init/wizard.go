package initcmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// RunWizard runs the interactive init wizard and returns the collected options.
func RunWizard(cwd string) (Options, error) {
	var opts Options
	askedForce := forceStep{}.Applicable(&opts, cwd)

	var fields []huh.Field
	for _, step := range Steps() {
		if !step.Applicable(&opts, cwd) {
			continue
		}
		fields = append(fields, step.Field(&opts))
	}

	if len(fields) > 0 {
		form := huh.NewForm(huh.NewGroup(fields...))
		if err := form.Run(); err != nil {
			return Options{}, fmt.Errorf("%w: %w", ErrAborted, err)
		}
	}

	if askedForce && !opts.Force {
		return Options{}, ErrAborted
	}

	return opts, nil
}

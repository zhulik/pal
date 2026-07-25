package initcmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// RunWizard runs the interactive init wizard and returns the collected options.
// Steps are shown one at a time; a declining answer that Abort reports stops
// the wizard immediately without asking later steps.
//
// opts may be pre-seeded (e.g. Module from a CLI argument); steps that are
// already satisfied are skipped via Applicable.
func RunWizard(cwd string, opts Options) (Options, error) {
	for _, step := range Steps() {
		if !step.Applicable(&opts, cwd) {
			continue
		}

		form := huh.NewForm(huh.NewGroup(step.Field(&opts)))
		if err := form.Run(); err != nil {
			return Options{}, fmt.Errorf("%w: %w", ErrAborted, err)
		}
		if step.Abort(&opts) {
			return Options{}, ErrAborted
		}
	}

	return opts, nil
}

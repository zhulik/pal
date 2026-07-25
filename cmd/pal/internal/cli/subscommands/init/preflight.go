package initcmd

import (
	"fmt"
	"os/exec"

	"github.com/zhulik/pal/cmd/pal/templates"
)

// preflight validates opts before any filesystem mutations. Call after the
// wizard (or after reading non-interactive flags) so template/git choices are
// final.
func preflight(opts Options) error {
	name := opts.Template
	if name == "" {
		name = defaultTemplate
	}
	if err := templates.Exists(name); err != nil {
		return err
	}

	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go is required on PATH: %w", err)
	}
	if opts.Git {
		if _, err := exec.LookPath("git"); err != nil {
			return fmt.Errorf("git is required on PATH (or pass --no-git): %w", err)
		}
	}
	return nil
}

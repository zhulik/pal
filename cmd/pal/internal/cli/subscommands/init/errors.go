package initcmd

import "errors"

var (
	// ErrAborted is returned when the user declines or cancels the wizard.
	ErrAborted = errors.New("init aborted")
	// ErrNotEmpty is returned when the target directory is not empty and force
	// was not set.
	ErrNotEmpty = errors.New("directory is not empty; re-run with --force to continue")
)

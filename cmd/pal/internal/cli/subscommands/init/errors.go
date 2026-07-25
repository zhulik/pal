package initcmd

import "errors"

var (
	// ErrAborted is returned when the user declines or cancels the wizard.
	ErrAborted = errors.New("init aborted")
	// ErrNotEmpty is returned when the target directory is not empty and force
	// was not set.
	ErrNotEmpty = errors.New("directory is not empty; re-run with --force to continue")
	// ErrModuleRequired is returned when a Go module path was not provided.
	ErrModuleRequired = errors.New("module path is required; pass it as: pal init <module>")
)

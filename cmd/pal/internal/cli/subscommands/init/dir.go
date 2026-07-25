package initcmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// IsEmpty reports whether dir contains no entries.
func IsEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("read directory: %w", err)
	}
	return len(entries) == 0, nil
}

// ResolveWorkDir picks where init should write files.
//
// If directory is empty, the process working directory is used.
// If directory exists and is a directory, it is used in place.
// If directory does not exist, a temp dir is created; promoteTo is set so the
// caller can copy the result to directory after a successful init. cleanup
// always removes the temp dir when one was created (call via defer).
func ResolveWorkDir(directory string) (workDir, promoteTo string, cleanup func(), err error) {
	cleanup = func() {}

	if directory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", cleanup, fmt.Errorf("get working directory: %w", err)
		}
		return cwd, "", cleanup, nil
	}

	abs, err := filepath.Abs(directory)
	if err != nil {
		return "", "", cleanup, fmt.Errorf("resolve directory: %w", err)
	}

	info, err := os.Stat(abs)
	if err == nil {
		if !info.IsDir() {
			return "", "", cleanup, fmt.Errorf("%s: not a directory", abs)
		}
		return abs, "", cleanup, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", "", cleanup, fmt.Errorf("stat directory: %w", err)
	}

	tmp, err := os.MkdirTemp("", "pal-init-*")
	if err != nil {
		return "", "", cleanup, fmt.Errorf("create temporary directory: %w", err)
	}
	cleanup = func() { _ = os.RemoveAll(tmp) }
	return tmp, abs, cleanup, nil
}

// PromoteWorkDir copies a successful init from tmp into target, creating any
// missing parent directories. The caller removes tmp via cleanup.
func PromoteWorkDir(tmp, target string) error {
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", target, err)
	}
	if err := os.CopyFS(target, os.DirFS(tmp)); err != nil {
		_ = os.RemoveAll(target)
		return fmt.Errorf("copy project to %s: %w", target, err)
	}
	return nil
}

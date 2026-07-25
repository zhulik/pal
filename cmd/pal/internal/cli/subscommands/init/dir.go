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

// ResolveWorkDir resolves the final project target and always creates a
// staging temp directory. Init must write only to workDir; on success call
// PromoteWorkDir(workDir, promoteTo, targetExisted). cleanup always removes
// the staging directory (call via defer).
//
// directory empty means the process working directory. If the target exists it
// must be a directory (emptiness is checked by the caller). If it is missing,
// targetExisted is false and promote creates it.
func ResolveWorkDir(directory string) (workDir, promoteTo string, targetExisted bool, cleanup func(), err error) {
	cleanup = func() {}

	promoteTo, targetExisted, err = resolveTarget(directory)
	if err != nil {
		return "", "", false, cleanup, err
	}

	tmp, err := os.MkdirTemp("", "pal-init-*")
	if err != nil {
		return "", "", false, cleanup, fmt.Errorf("create temporary directory: %w", err)
	}
	cleanup = func() { _ = os.RemoveAll(tmp) }
	return tmp, promoteTo, targetExisted, cleanup, nil
}

func resolveTarget(directory string) (target string, existed bool, err error) {
	if directory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false, fmt.Errorf("get working directory: %w", err)
		}
		return cwd, true, nil
	}

	abs, err := filepath.Abs(directory)
	if err != nil {
		return "", false, fmt.Errorf("resolve directory: %w", err)
	}

	info, err := os.Stat(abs)
	if err == nil {
		if !info.IsDir() {
			return "", false, fmt.Errorf("%s: not a directory", abs)
		}
		return abs, true, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", false, fmt.Errorf("stat directory: %w", err)
	}
	return abs, false, nil
}

// PromoteWorkDir copies a successful init from stage into target.
//
// If targetExisted is false, missing parents are created; on copy failure the
// newly created target tree is removed. If targetExisted is true, target must
// already be an empty directory; on copy failure its contents are cleared so
// it remains an empty directory (the directory itself is not removed).
func PromoteWorkDir(stage, target string, targetExisted bool) error {
	if !targetExisted {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", target, err)
		}
	}
	if err := os.CopyFS(target, os.DirFS(stage)); err != nil {
		if targetExisted {
			_ = clearDir(target)
		} else {
			_ = os.RemoveAll(target)
		}
		return fmt.Errorf("copy project to %s: %w", target, err)
	}
	return nil
}

func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

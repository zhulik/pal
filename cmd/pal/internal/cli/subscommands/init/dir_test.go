package initcmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	initcmd "github.com/zhulik/pal/cmd/pal/internal/cli/subscommands/init"
)

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		empty, err := initcmd.IsEmpty(dir)
		require.NoError(t, err)
		require.True(t, empty)
	})

	t.Run("non-empty directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		empty, err := initcmd.IsEmpty(dir)
		require.NoError(t, err)
		require.False(t, empty)
	})

	t.Run("missing directory", func(t *testing.T) {
		t.Parallel()
		_, err := initcmd.IsEmpty(filepath.Join(t.TempDir(), "does-not-exist"))
		require.Error(t, err)
	})
}

//nolint:paralleltest // t.Chdir cannot be combined with t.Parallel
func TestResolveWorkDirEmptyUsesCwdAsPromoteTarget(t *testing.T) {
	t.Chdir(t.TempDir())

	workDir, promoteTo, targetExisted, cleanup, err := initcmd.ResolveWorkDir("")
	require.NoError(t, err)
	defer cleanup()
	got, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, got, promoteTo)
	require.True(t, targetExisted)
	require.NotEqual(t, promoteTo, workDir)
	require.DirExists(t, workDir)
}

func TestResolveWorkDir(t *testing.T) {
	t.Parallel()

	t.Run("existing directory stages in temp", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		workDir, promoteTo, targetExisted, cleanup, err := initcmd.ResolveWorkDir(dir)
		require.NoError(t, err)
		defer cleanup()
		require.Equal(t, dir, promoteTo)
		require.True(t, targetExisted)
		require.NotEqual(t, dir, workDir)
		require.DirExists(t, workDir)
	})

	t.Run("existing file errors", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		file := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))

		_, _, _, cleanup, err := initcmd.ResolveWorkDir(file)
		cleanup()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not a directory")
	})

	t.Run("missing path uses temp and promote target", func(t *testing.T) {
		t.Parallel()
		parent := t.TempDir()
		target := filepath.Join(parent, "new", "app")

		workDir, promoteTo, targetExisted, cleanup, err := initcmd.ResolveWorkDir(target)
		require.NoError(t, err)
		defer cleanup()
		require.NotEqual(t, target, workDir)
		require.DirExists(t, workDir)
		require.Equal(t, target, promoteTo)
		require.False(t, targetExisted)
	})
}

func TestPromoteWorkDir(t *testing.T) {
	t.Parallel()

	t.Run("creates missing target", func(t *testing.T) {
		t.Parallel()
		stage := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(stage, "go.mod"), []byte("module x\n"), 0o644))
		target := filepath.Join(t.TempDir(), "nested", "app")

		require.NoError(t, initcmd.PromoteWorkDir(stage, target, false))
		require.FileExists(t, filepath.Join(target, "go.mod"))
		require.DirExists(t, stage)
	})

	t.Run("copies into existing empty target", func(t *testing.T) {
		t.Parallel()
		stage := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(stage, "go.mod"), []byte("module x\n"), 0o644))
		target := t.TempDir()

		require.NoError(t, initcmd.PromoteWorkDir(stage, target, true))
		require.FileExists(t, filepath.Join(target, "go.mod"))
	})

	t.Run("rejects non-empty existing target without clearing it", func(t *testing.T) {
		t.Parallel()
		stage := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(stage, "go.mod"), []byte("module x\n"), 0o644))
		target := t.TempDir()
		keep := filepath.Join(target, "keep.txt")
		require.NoError(t, os.WriteFile(keep, []byte("mine"), 0o644))

		err := initcmd.PromoteWorkDir(stage, target, true)
		require.ErrorIs(t, err, initcmd.ErrNotEmpty)
		data, err := os.ReadFile(keep)
		require.NoError(t, err)
		require.Equal(t, "mine", string(data))
		_, err = os.Stat(filepath.Join(target, "go.mod"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}

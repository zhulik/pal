package initcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	t.Parallel()

	const module = "example.com/myapp"

	t.Run("empty directory succeeds", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := (&runner{opts: Options{NoInteractive: true, Directory: dir, Module: module, Template: "cli"}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir)
		requireCLITemplate(t, dir)
	})

	t.Run("non-empty directory fails", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))

		err := (&runner{opts: Options{NoInteractive: true, Directory: dir, Module: module, Template: "cli"}}).Run(t.Context())
		require.ErrorIs(t, err, ErrNotEmpty)
	})

	t.Run("missing module fails", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := (&runner{opts: Options{NoInteractive: true, Directory: dir, Template: "cli"}}).Run(t.Context())
		require.ErrorIs(t, err, ErrModuleRequired)
	})

	t.Run("git init when Git is true", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := (&runner{opts: Options{NoInteractive: true, Directory: dir, Module: module, Template: "cli", Git: true}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir)
		requireCLITemplate(t, dir)

		info, err := os.Stat(filepath.Join(dir, ".git"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("skips git init when Git is false", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := (&runner{opts: Options{NoInteractive: true, Directory: dir, Module: module, Template: "cli", Git: false}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir)
		requireCLITemplate(t, dir)

		_, err = os.Stat(filepath.Join(dir, ".git"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("unknown template fails", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := (&runner{opts: Options{NoInteractive: true, Directory: dir, Module: module, Template: "nope"}}).Run(t.Context())
		require.Error(t, err)
		require.Contains(t, err.Error(), "nope")
	})

	t.Run("directory flag uses existing empty dir", func(t *testing.T) {
		t.Parallel()
		target := t.TempDir()

		err := (&runner{opts: Options{
			NoInteractive: true,
			Directory:     target,
			Module:        module,
			Template:      "cli",
		}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, target)
		requireCLITemplate(t, target)
	})

	t.Run("directory flag creates missing nested path", func(t *testing.T) {
		t.Parallel()
		cwd := t.TempDir()
		target := filepath.Join(cwd, "new", "nested", "app")

		err := (&runner{opts: Options{
			NoInteractive: true,
			Directory:     target,
			Module:        module,
			Template:      "cli",
		}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, target)
		requireCLITemplate(t, target)
	})

	t.Run("directory flag failure leaves missing path absent", func(t *testing.T) {
		t.Parallel()
		cwd := t.TempDir()
		target := filepath.Join(cwd, "missing", "app")

		err := (&runner{opts: Options{
			NoInteractive: true,
			Directory:     target,
			Module:        module,
			Template:      "nope",
		}}).Run(t.Context())
		require.Error(t, err)
		_, err = os.Stat(target)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("directory flag non-empty fails", func(t *testing.T) {
		t.Parallel()
		target := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(target, "file.txt"), []byte("x"), 0o644))

		err := (&runner{opts: Options{
			NoInteractive: true,
			Directory:     target,
			Module:        module,
			Template:      "cli",
		}}).Run(t.Context())
		require.ErrorIs(t, err, ErrNotEmpty)
	})
}

func requireGoMod(t *testing.T, dir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	require.NoError(t, err)
	require.Contains(t, string(data), "module example.com/myapp")
}

func requireCLITemplate(t *testing.T, dir string) {
	t.Helper()
	mainPath := filepath.Join(dir, "cmd", "myapp", "main.go")
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "package main")
	require.Contains(t, string(data), `"myapp running"`)
	require.Contains(t, string(data), `"github.com/zhulik/pal"`)
	require.NotContains(t, string(data), "{{.Package}}")
}

func TestOptionsArgs(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"--no-git"}, Options{}.Args())
	require.Equal(t, []string{"example.com/app"}, Options{Module: "example.com/app", Git: true}.Args())
	require.Equal(t, []string{"example.com/app", "-d", "./app"}, Options{Module: "example.com/app", Directory: "./app", Git: true}.Args())
	require.Equal(t, []string{"example.com/app", "--template", "other"}, Options{Module: "example.com/app", Template: "other", Git: true}.Args())
}

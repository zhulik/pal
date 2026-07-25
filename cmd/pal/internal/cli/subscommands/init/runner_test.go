package initcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	const module = "example.com/myapp"

	t.Run("empty directory succeeds without force", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module, Template: "cli"}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)
		requireCLITemplate(t, dir, "myapp")
	})

	t.Run("non-empty without force fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module, Template: "cli"}}).Run(t.Context())
		require.ErrorIs(t, err, ErrNotEmpty)
	})

	t.Run("non-empty with force succeeds", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Force: true, Module: module, Template: "cli"}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)
		requireCLITemplate(t, dir, "myapp")
	})

	t.Run("missing module fails", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Template: "cli"}}).Run(t.Context())
		require.ErrorIs(t, err, ErrModuleRequired)
	})

	t.Run("git init when Git is true", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module, Template: "cli", Git: true}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)
		requireCLITemplate(t, dir, "myapp")

		info, err := os.Stat(filepath.Join(dir, ".git"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("skips git init when Git is false", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module, Template: "cli", Git: false}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)
		requireCLITemplate(t, dir, "myapp")

		_, err = os.Stat(filepath.Join(dir, ".git"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("unknown template fails", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module, Template: "nope"}}).Run(t.Context())
		require.Error(t, err)
		require.Contains(t, err.Error(), "nope")
	})
}

func requireGoMod(t *testing.T, dir, module string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	require.NoError(t, err)
	require.Contains(t, string(data), "module "+module)
}

func requireCLITemplate(t *testing.T, dir, pkg string) {
	t.Helper()
	mainPath := filepath.Join(dir, "cmd", pkg, "main.go")
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "package main")
	require.Contains(t, string(data), `"`+pkg+` running"`)
	require.Contains(t, string(data), `"github.com/zhulik/pal"`)
	require.NotContains(t, string(data), "{{.Package}}")
}

func TestOptionsArgs(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"--no-git"}, Options{}.Args())
	require.Equal(t, []string{"example.com/app"}, Options{Module: "example.com/app", Git: true}.Args())
	require.Equal(t, []string{"example.com/app", "--force"}, Options{Module: "example.com/app", Force: true, Git: true}.Args())
	require.Equal(t, []string{"example.com/app", "--force", "--no-git"}, Options{Module: "example.com/app", Force: true}.Args())
	require.Equal(t, []string{"example.com/app", "--template", "other"}, Options{Module: "example.com/app", Template: "other", Git: true}.Args())
}

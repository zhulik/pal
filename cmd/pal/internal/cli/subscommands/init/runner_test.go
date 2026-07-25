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

		err := (&runner{opts: Options{NoInteractive: true, Module: module}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)
	})

	t.Run("non-empty without force fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module}}).Run(t.Context())
		require.ErrorIs(t, err, ErrNotEmpty)
	})

	t.Run("non-empty with force succeeds", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Force: true, Module: module}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)
	})

	t.Run("missing module fails", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true}}).Run(t.Context())
		require.ErrorIs(t, err, ErrModuleRequired)
	})

	t.Run("git init when Git is true", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module, Git: true}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)

		info, err := os.Stat(filepath.Join(dir, ".git"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("skips git init when Git is false", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Module: module, Git: false}}).Run(t.Context())
		require.NoError(t, err)
		requireGoMod(t, dir, module)

		_, err = os.Stat(filepath.Join(dir, ".git"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}

func requireGoMod(t *testing.T, dir, module string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	require.NoError(t, err)
	require.Contains(t, string(data), "module "+module)
}

func TestOptionsArgs(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"--no-git"}, Options{}.Args())
	require.Equal(t, []string{"example.com/app"}, Options{Module: "example.com/app", Git: true}.Args())
	require.Equal(t, []string{"example.com/app", "--force"}, Options{Module: "example.com/app", Force: true, Git: true}.Args())
	require.Equal(t, []string{"example.com/app", "--force", "--no-git"}, Options{Module: "example.com/app", Force: true}.Args())
}

package initcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	t.Run("empty directory succeeds without force", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true}}).Run(t.Context())
		require.NoError(t, err)
	})

	t.Run("non-empty without force fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true}}).Run(t.Context())
		require.ErrorIs(t, err, ErrNotEmpty)
	})

	t.Run("non-empty with force succeeds", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Force: true}}).Run(t.Context())
		require.NoError(t, err)
	})

	t.Run("git init when Git is true", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Git: true}}).Run(t.Context())
		require.NoError(t, err)

		info, err := os.Stat(filepath.Join(dir, ".git"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("skips git init when Git is false", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := (&runner{opts: Options{NoInteractive: true, Git: false}}).Run(t.Context())
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(dir, ".git"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestOptionsArgs(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"--no-git"}, Options{}.Args())
	require.Empty(t, Options{Git: true}.Args())
	require.Equal(t, []string{"--force"}, Options{Force: true, Git: true}.Args())
	require.Equal(t, []string{"--force", "--no-git"}, Options{Force: true}.Args())
}

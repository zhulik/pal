package initcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		empty, err := IsEmpty(dir)
		require.NoError(t, err)
		require.True(t, empty)
	})

	t.Run("non-empty directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644))
		empty, err := IsEmpty(dir)
		require.NoError(t, err)
		require.False(t, empty)
	})

	t.Run("missing directory", func(t *testing.T) {
		t.Parallel()
		_, err := IsEmpty(filepath.Join(t.TempDir(), "does-not-exist"))
		require.Error(t, err)
	})
}

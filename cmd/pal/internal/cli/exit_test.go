package cli_test

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBinary_ExitCodes(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	bin := buildPalBinary(t, root)

	t.Run("version exits 0", func(t *testing.T) {
		t.Parallel()
		cmd := exec.CommandContext(t.Context(), bin, "version")
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "pal version\n%s", out)
	})

	t.Run("unknown template exits 1", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		cmd := exec.CommandContext(t.Context(), bin,
			"init", "--no-interactive", "-d", dir,
			"--template", "nope",
			"example.com/paltest",
		)
		out, err := cmd.CombinedOutput()
		require.Error(t, err)
		requireExitCode(t, err, 1)
		require.Contains(t, string(out), "nope")
		requireEmptyDir(t, dir)
	})

	t.Run("missing module exits 1", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		cmd := exec.CommandContext(t.Context(), bin,
			"init", "--no-interactive", "-d", dir,
			"--template", "cli", "--no-git",
		)
		out, err := cmd.CombinedOutput()
		require.Error(t, err)
		requireExitCode(t, err, 1)
		require.Contains(t, string(out), "module path is required")
		requireEmptyDir(t, dir)
	})
}

func requireExitCode(t *testing.T, err error, want int) {
	t.Helper()
	var ee *exec.ExitError
	require.Truef(t, errors.As(err, &ee), "want ExitError, got %T: %v", err, err)
	require.Equal(t, want, ee.ExitCode())
}

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries)
}

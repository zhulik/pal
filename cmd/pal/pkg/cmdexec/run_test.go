package cmdexec_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zhulik/pal/cmd/pal/pkg/cmdexec"
)

func TestRunSuccess(t *testing.T) {
	t.Parallel()
	require.NoError(t, cmdexec.Run(t.Context(), "", "true"))
}

func TestRunFailure(t *testing.T) {
	t.Parallel()
	err := cmdexec.Run(t.Context(), "", "false")
	require.Error(t, err)

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
}

func TestRunDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, cmdexec.Run(t.Context(), dir, "pwd"))
}

package cmdexec

import (
	"context"
	"os"
	"os/exec"
)

// Run runs name with args under ctx, forwarding stdout and stderr to the
// current process in real time. Prefer this over [*exec.Cmd.CombinedOutput] or
// [*exec.Cmd.Output] when the user should see subprocess progress as it happens.
//
// dir is the working directory for the command; empty means the current process
// directory. When Stdout/Stderr are [*os.File] (as here), os/exec connects the
// file descriptors directly — no extra buffering or copy goroutines.
func Run(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

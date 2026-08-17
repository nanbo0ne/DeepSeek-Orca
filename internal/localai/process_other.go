//go:build !windows

package localai

import (
	"os/exec"
)

func configureHiddenProcess(*exec.Cmd) {}

func terminateProcess(cmd *exec.Cmd, _ <-chan struct{}) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

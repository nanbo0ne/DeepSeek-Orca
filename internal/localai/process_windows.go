//go:build windows

package localai

import (
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

func configureHiddenProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW | windows.CREATE_NEW_PROCESS_GROUP}
}

func terminateProcess(cmd *exec.Cmd, done <-chan struct{}) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Signal(syscall.Signal(windows.CTRL_BREAK_EVENT))
	select {
	case <-done:
		return nil
	case <-time.After(1200 * time.Millisecond):
		if err := cmd.Process.Kill(); err != nil && err != windows.ERROR_ACCESS_DENIED {
			return err
		}
		return nil
	}
}

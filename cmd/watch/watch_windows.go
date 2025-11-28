//go:build windows

package watch

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func (p *RealProcessRunner) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	// Use taskkill to kill process tree
	// /F = force
	// /T = tree (child processes)
	// /PID = process id
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(p.cmd.Process.Pid))
	// We don't necessarily want to pipe stdout/stderr for the kill command itself unless debugging
	// but keeping it consistent might be useful. For now, let's suppress it to avoid noise,
	// or maybe just ignore the error if the process is already dead.
	return killCmd.Run()
}

func NewProcessRunner(params *Params) func() ProcessRunner {
	return func() ProcessRunner {
		c := exec.Command("cmd", "/C", params.Execute)
		// Create a new process group
		c.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return &RealProcessRunner{cmd: c}
	}
}

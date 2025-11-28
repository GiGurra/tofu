//go:build !windows

package watch

import (
	"os"
	"os/exec"
	"syscall"
)

func (p *RealProcessRunner) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	// Try to kill the process group first (kills child processes too)
	if err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL); err != nil {
		// Fallback to killing just the process
		return p.cmd.Process.Kill()
	}
	return nil
}

func NewProcessRunner(params *Params) func() ProcessRunner {
	return func() ProcessRunner {
		c := exec.Command("sh", "-c", params.Execute)
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return &RealProcessRunner{cmd: c}
	}
}

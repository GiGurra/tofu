package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

func (p *RealProcessRunner) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

func NewProcessRunner(params *WatchParams) func() ProcessRunner {
	return func() ProcessRunner {
		c := exec.Command("sh", "-c", params.Execute)
		c.SysProcAttr = &syscall.SysProcAttr{ /*Setpgid: true*/ }
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return &RealProcessRunner{cmd: c}
	}
}

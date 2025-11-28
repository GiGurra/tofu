package port

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestHelperProcess is not a real test. It's a helper process invoked by TestPortCommand.
// It starts a TCP listener on a random port and prints the port number to stdout.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	// Print the port so the parent knows. Use a prefix to make it easy to find.
	fmt.Printf("PORT:%d\n", listener.Addr().(*net.TCPAddr).Port)

	// Block until killed
	select {}
}

func TestPortCommand(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip("Test does not work on windows")
	}

	// Find the test executable path (ourselves)
	exe, err := os.Executable()
	if err != nil {
		t.Skip("Could not find executable path, skipping test")
	}

	// Start the helper process
	cmd := exec.Command(exe, "-test.run=TestHelperProcess", "-test.v")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	// Capture stdout to get the port
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	t.Logf("Running helper: %s -test.run=TestHelperProcess", exe)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start helper process: %v", err)
	}

	// Wait a bit for it to start and print the port
	port := 0
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if outBuf.Len() > 0 {
			lines := strings.Split(outBuf.String(), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PORT:") {
					pStr := strings.TrimPrefix(line, "PORT:")
					p, err := strconv.Atoi(strings.TrimSpace(pStr))
					if err == nil {
						port = p
						break
					}
				}
			}
			if port > 0 {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	if port == 0 {
		cmd.Process.Kill()
		t.Fatalf("Helper process did not report a port. Stderr: %s. Stdout: %s", errBuf.String(), outBuf.String())
	}

	t.Logf("Helper process listening on port %d (PID %d)", port, cmd.Process.Pid)

	// 1. Test Listing
	// We expect to find this port
	listStdout := &bytes.Buffer{}
	listStderr := &bytes.Buffer{}
	params := &Params{
		PortNum: port,
	}

	if err := Run(params, listStdout, listStderr); err != nil {
		t.Errorf("Run failed to list port %d: %v. Stderr: %s", port, err, listStderr.String())
	}
	// Check if the stdout contains the port and PID
	if !strings.Contains(listStdout.String(), strconv.Itoa(port)) || !strings.Contains(listStdout.String(), strconv.Itoa(int(cmd.Process.Pid))) {
		t.Errorf("Expected list output to contain port %d and PID %d, got:\n%s", port, cmd.Process.Pid, listStdout.String())
	}

	// 2. Test Killing
	killStdout := &bytes.Buffer{}
	killStderr := &bytes.Buffer{}
	paramsKill := &Params{
		PortNum: port,
		Kill:    true,
	}

	if err := Run(paramsKill, killStdout, killStderr); err != nil {
		t.Errorf("Run failed to kill port %d: %v. Stderr: %s", port, err, killStderr.String())
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		t.Log("Helper process exited successfully")
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		t.Errorf("Process did not exit after Run --kill")
	}
}

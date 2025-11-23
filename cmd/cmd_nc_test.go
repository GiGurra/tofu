package cmd

import (
	"bytes"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunNc_ClientServer_TCP(t *testing.T) {
	// Find a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	serverStdout := &bytes.Buffer{}
	clientStdout := &bytes.Buffer{}
	var serverStderr, clientStderr bytes.Buffer

	serverStdin := strings.NewReader("Hello Client")
	clientStdin := strings.NewReader("Hello Server")

	portStr := strconv.Itoa(port)

	// Wait a bit for server to start listening
	time.Sleep(500 * time.Millisecond)

	// Run Server
	go func() {
		defer wg.Done()
		params := &NcParams{
			Args:   []string{portStr},
			Listen: true,
		}
		// Give the server a moment to fail if port is taken, but we closed it.
		// However, in this test, runNc blocks.
		// We expect runNc to return when connection closes.
		err := runNc(params, serverStdin, serverStdout, &serverStderr)
		if err != nil {
			t.Logf("Server exited with error (might be expected on close): %v", err)
		}
	}()

	// Wait a bit for server to start listening
	time.Sleep(500 * time.Millisecond)

	// Run Client
	go func() {
		defer wg.Done()
		params := &NcParams{
			Args: []string{"127.0.0.1", portStr},
		}
		err := runNc(params, clientStdin, clientStdout, &clientStderr)
		if err != nil {
			t.Logf("Client exited with error: %v", err)
		}
	}()

	// We need a way to stop them. runNc blocks until connection closes.
	// The client sends "Hello Server" and then closes its write side (in pipeStream).
	// The server receives EOF, loops, and should finish?
	// pipeStream implementation:
	//   go copy(conn, stdin) -> sends stdin then closes write.
	//   go copy(stdout, conn) -> reads until EOF.
	//
	// When client finishes sending "Hello Server", it closes write.
	// Server sees EOF from conn. copy(stdout, conn) finishes.
	// Server pipeStream should return.
	//
	// However, server also sends "Hello Client".
	// Client needs to read it.

	// Let's rely on timeouts if it hangs, but in a test we want it to finish cleanly.

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out")
	}

	// Verify output
	if !strings.Contains(serverStdout.String(), "Hello Server") {
		t.Errorf("Server didn't receive expected message. Got: %q", serverStdout.String())
	}
	if !strings.Contains(clientStdout.String(), "Hello Client") {
		t.Errorf("Client didn't receive expected message. Got: %q", clientStdout.String())
	}
}

func TestParseNcArgs(t *testing.T) {
	tests := []struct {
		args     []string
		listen   bool
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{[]string{"localhost", "8080"}, false, "localhost", "8080", false},
		{[]string{"8080"}, true, "", "8080", false},
		{[]string{"127.0.0.1:8080"}, false, "127.0.0.1", "8080", false},
		{[]string{"8080"}, false, "127.0.0.1", "8080", false},
		{[]string{}, false, "", "", true},
	}

	for _, tt := range tests {
		host, port, err := parseNcArgs(tt.args, tt.listen)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseNcArgs(%v, %v) error = %v, wantErr %v", tt.args, tt.listen, err, tt.wantErr)
			continue
		}
		if host != tt.wantHost {
			t.Errorf("parseNcArgs(%v, %v) host = %v, want %v", tt.args, tt.listen, host, tt.wantHost)
		}
		if port != tt.wantPort {
			t.Errorf("parseNcArgs(%v, %v) port = %v, want %v", tt.args, tt.listen, port, tt.wantPort)
		}
	}
}

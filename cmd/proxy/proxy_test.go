package proxy

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// startEchoServer starts a TCP server that echoes back everything it receives.
// Returns the listener (caller must close).
func startEchoServer(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start echo server: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, conn)
			}()
		}
	}()
	return ln
}

// freePort finds a free TCP port by briefly listening and closing.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

// startProxy starts the proxy in a goroutine and waits for it to be accepting.
func startProxy(t *testing.T, params *Params) {
	t.Helper()
	go run(params)
	// Wait for proxy to start accepting
	for range 50 {
		conn, err := net.DialTimeout("tcp", params.Listen, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("proxy did not start listening on %s", params.Listen)
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.input)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCopyWithIdleTimeout_NoTimeout(t *testing.T) {
	server, client := net.Pipe()
	dst, dstClient := net.Pipe()
	defer server.Close()
	defer client.Close()
	defer dst.Close()
	defer dstClient.Close()

	msg := "hello world"
	go func() {
		client.Write([]byte(msg))
		client.Close()
	}()
	go io.Copy(io.Discard, dstClient)

	n := copyWithIdleTimeout(dst, server, 0)
	if n != int64(len(msg)) {
		t.Errorf("copyWithIdleTimeout returned %d bytes, want %d", n, len(msg))
	}
}

func TestCopyWithIdleTimeout_WithTimeout(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Write some data, then go idle
	msg := "hello"
	go func() {
		client.Write([]byte(msg))
		// Don't close - let idle timeout trigger
	}()

	dst, dstClient := net.Pipe()
	defer dst.Close()
	defer dstClient.Close()

	// Read what arrives at dst
	go io.Copy(io.Discard, dstClient)

	start := time.Now()
	n := copyWithIdleTimeout(dst, server, 200*time.Millisecond)
	elapsed := time.Since(start)

	if n != int64(len(msg)) {
		t.Errorf("copied %d bytes, want %d", n, len(msg))
	}
	if elapsed < 150*time.Millisecond || elapsed > 1*time.Second {
		t.Errorf("idle timeout took %v, expected ~200ms", elapsed)
	}
}

func TestDialWithRetry_Success(t *testing.T) {
	echo := startEchoServer(t)
	defer echo.Close()

	params := &Params{
		Target:         echo.Addr().String(),
		ConnectTimeout: 1000,
		Retries:        0,
	}

	conn, err := dialWithRetry(params, 1)
	if err != nil {
		t.Fatalf("dialWithRetry failed: %v", err)
	}
	conn.Close()
}

func TestDialWithRetry_FailNoRetry(t *testing.T) {
	// Target that doesn't exist - use a port that's definitely closed
	port := freePort(t)
	params := &Params{
		Target:         fmt.Sprintf("127.0.0.1:%d", port),
		ConnectTimeout: 200,
		Retries:        0,
	}

	_, err := dialWithRetry(params, 1)
	if err == nil {
		t.Fatal("expected error connecting to closed port, got nil")
	}
}

func TestDialWithRetry_RetriesAndSucceeds(t *testing.T) {
	port := freePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	// Start listener after a short delay (simulating target coming up)
	go func() {
		time.Sleep(300 * time.Millisecond)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return
		}
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	params := &Params{
		Target:         addr,
		ConnectTimeout: 200,
		Retries:        5,
		RetryInterval:  100,
		Verbose:        true,
	}

	conn, err := dialWithRetry(params, 1)
	if err != nil {
		t.Fatalf("dialWithRetry failed after retries: %v", err)
	}
	conn.Close()
}

func TestDialWithRetry_ExhaustsRetries(t *testing.T) {
	port := freePort(t)
	params := &Params{
		Target:         fmt.Sprintf("127.0.0.1:%d", port),
		ConnectTimeout: 100,
		Retries:        2,
		RetryInterval:  50,
	}

	start := time.Now()
	_, err := dialWithRetry(params, 1)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// Should have taken at least 2 retry intervals (50ms * 2 = 100ms)
	if elapsed < 80*time.Millisecond {
		t.Errorf("retries completed too fast (%v), expected at least ~100ms", elapsed)
	}
}

func TestProxyBasic(t *testing.T) {
	echo := startEchoServer(t)
	defer echo.Close()

	proxyPort := freePort(t)
	params := &Params{
		Listen:         fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Target:         echo.Addr().String(),
		ConnectTimeout: 2000,
	}
	startProxy(t, params)

	// Connect through proxy and send data
	conn, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("failed to connect through proxy: %v", err)
	}
	defer conn.Close()

	msg := "hello through proxy"
	conn.Write([]byte(msg))
	conn.(*net.TCPConn).CloseWrite()

	got, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if string(got) != msg {
		t.Errorf("got %q, want %q", string(got), msg)
	}
}

func TestProxyLargeData(t *testing.T) {
	echo := startEchoServer(t)
	defer echo.Close()

	proxyPort := freePort(t)
	params := &Params{
		Listen:         fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Target:         echo.Addr().String(),
		ConnectTimeout: 2000,
	}
	startProxy(t, params)

	conn, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("failed to connect through proxy: %v", err)
	}
	defer conn.Close()

	// Send 1MB of data
	data := strings.Repeat("x", 1024*1024)
	go func() {
		conn.Write([]byte(data))
		conn.(*net.TCPConn).CloseWrite()
	}()

	got, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if len(got) != len(data) {
		t.Errorf("got %d bytes, want %d", len(got), len(data))
	}
}

func TestProxyMultipleConnections(t *testing.T) {
	echo := startEchoServer(t)
	defer echo.Close()

	proxyPort := freePort(t)
	params := &Params{
		Listen:         fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Target:         echo.Addr().String(),
		ConnectTimeout: 2000,
	}
	startProxy(t, params)

	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", params.Listen, time.Second)
			if err != nil {
				t.Errorf("[%d] connect failed: %v", id, err)
				return
			}
			defer conn.Close()

			msg := fmt.Sprintf("hello from %d", id)
			conn.Write([]byte(msg))
			conn.(*net.TCPConn).CloseWrite()

			got, err := io.ReadAll(conn)
			if err != nil {
				t.Errorf("[%d] read failed: %v", id, err)
				return
			}
			if string(got) != msg {
				t.Errorf("[%d] got %q, want %q", id, string(got), msg)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("test timed out")
	}
}

func TestProxyMaxConnections(t *testing.T) {
	// Slow echo server that holds connections open
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start slow server: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				// Hold connection open, echo slowly
				io.Copy(conn, conn)
			}()
		}
	}()

	proxyPort := freePort(t)
	params := &Params{
		Listen:         fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Target:         ln.Addr().String(),
		ConnectTimeout: 2000,
		MaxConns:       2,
	}
	startProxy(t, params)

	// Fill up the 2 allowed slots
	conn1, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("conn1 failed: %v", err)
	}
	defer conn1.Close()

	conn2, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("conn2 failed: %v", err)
	}
	defer conn2.Close()

	// Give proxy time to process both connections
	time.Sleep(100 * time.Millisecond)

	// Third connection should be accepted at TCP level but get closed by proxy
	conn3, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("conn3 dial failed: %v", err)
	}
	defer conn3.Close()

	// The proxy will close conn3 immediately; reading should return EOF or error
	conn3.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 1)
	_, err = conn3.Read(buf)
	if err == nil {
		t.Error("expected conn3 to be closed by proxy, but read succeeded")
	}
}

func TestProxyIdleTimeout(t *testing.T) {
	echo := startEchoServer(t)
	defer echo.Close()

	proxyPort := freePort(t)
	params := &Params{
		Listen:         fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Target:         echo.Addr().String(),
		ConnectTimeout: 2000,
		IdleTimeout:    200,
	}
	startProxy(t, params)

	conn, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	// Send some data, verify echo works
	conn.Write([]byte("hi"))
	buf := make([]byte, 10)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(buf[:n]) != "hi" {
		t.Errorf("got %q, want %q", string(buf[:n]), "hi")
	}

	// Now go idle - connection should close after ~200ms
	start := time.Now()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buf)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected connection to close due to idle timeout")
	}
	if elapsed < 150*time.Millisecond || elapsed > 1*time.Second {
		t.Errorf("idle timeout took %v, expected ~200ms", elapsed)
	}
}

func TestProxyTargetDown(t *testing.T) {
	// Proxy to a port nothing listens on
	targetPort := freePort(t)
	proxyPort := freePort(t)
	params := &Params{
		Listen:         fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Target:         fmt.Sprintf("127.0.0.1:%d", targetPort),
		ConnectTimeout: 200,
		Retries:        0,
	}
	startProxy(t, params)

	conn, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("connect to proxy failed: %v", err)
	}
	defer conn.Close()

	// Proxy should close our connection since it can't reach target
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err == nil {
		t.Error("expected connection to close when target is down")
	}
}

func TestProxyVerbose(t *testing.T) {
	echo := startEchoServer(t)
	defer echo.Close()

	proxyPort := freePort(t)
	params := &Params{
		Listen:         fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Target:         echo.Addr().String(),
		ConnectTimeout: 2000,
		Verbose:        true,
	}
	startProxy(t, params)

	// Just verify it doesn't panic with verbose on
	conn, err := net.DialTimeout("tcp", params.Listen, time.Second)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	conn.Write([]byte("test"))
	conn.(*net.TCPConn).CloseWrite()
	io.ReadAll(conn)
	conn.Close()

	// Small delay to let verbose output print
	time.Sleep(100 * time.Millisecond)
}

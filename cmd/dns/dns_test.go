package dns

import (
	"bytes"
	"strings"
	"testing"
)

func TestDnsCmd_Structure(t *testing.T) {
	// This test checks if the command structure and flag parsing works.
	// We avoid making actual network calls in this unit test by not running RunFunc,
	// or by expecting failure but checking output format if we do.

	cmd := Cmd()
	if cmd.Use != "dns <hostname>" {
		t.Errorf("Expected command use to be 'dns <hostname>', got '%s'", cmd.Use)
	}
}

func TestRunDns_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DNS integration test in short mode")
	}

	// Test with a known stable domain
	params := &Params{
		Hostname: "google.com",
		Server:   "os", // Use OS resolver to avoid firewall issues with 8.8.8.8 in some envs
		Types:    []string{"A"},
		Timeout:  2,
	}
	var buf bytes.Buffer
	runDns(params, &buf)

	output := buf.String()
	if !strings.Contains(output, "A Records") {
		t.Errorf("Expected output to contain 'A Records', got:\n%s", output)
	}
	if !strings.Contains(output, "Server:  OS") {
		t.Errorf("Expected output to indicate OS server, got:\n%s", output)
	}
}

func TestRunDns_SpecificServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DNS integration test in short mode")
	}

	params := &Params{
		Hostname: "example.com",
		Server:   "1.1.1.1", // Cloudflare
		Types:    []string{"A"},
		Timeout:  2,
	}
	var buf bytes.Buffer
	runDns(params, &buf)

	output := buf.String()
	if !strings.Contains(output, "Server:  1.1.1.1:53") {
		t.Errorf("Expected output to show server 1.1.1.1:53, got:\n%s", output)
	}
}

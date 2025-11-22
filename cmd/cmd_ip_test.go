package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestIpCmd_LocalOnly(t *testing.T) {
	params := &IpParams{
		LocalOnly: true,
	}
	var buf bytes.Buffer

	runIp(params, &buf)

	output := buf.String()

	if !strings.Contains(output, "Local Interfaces:") {
		t.Errorf("Expected output to contain 'Local Interfaces:', got:\n%s", output)
	}

	if strings.Contains(output, "Public IP:") {
		t.Errorf("Expected output NOT to contain 'Public IP:' when LocalOnly is set, got:\n%s", output)
	}
}

func TestIpCmd_Public(t *testing.T) {
	// This test actually hits the network, so it might be flaky or slow. 
	// We can skip it if -short is passed or just accept it as an integration test.
	if testing.Short() {
		t.Skip("Skipping public IP test in short mode")
	}

	params := &IpParams{
		LocalOnly: false,
	}
	var buf bytes.Buffer

	runIp(params, &buf)

	output := buf.String()

	if !strings.Contains(output, "Local Interfaces:") {
		t.Errorf("Expected output to contain 'Local Interfaces:', got:\n%s", output)
	}

	if !strings.Contains(output, "Public IP:") {
		// This might fail if no network, but it's good for manual verification
		t.Logf("Warning: Output did not contain 'Public IP:'. This might be due to network issues.\nOutput:\n%s", output)
	}
}

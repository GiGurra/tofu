package free

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/shirou/gopsutil/v3/mem"
)

// Mocking gopsutil for testing
// Note: This requires a bit more effort if we want to truly mock
// the underlying functions. For now, we will rely on a happy path test
// and check output formatting.

func TestFreeCommand(t *testing.T) {
	tests := []struct {
		name       string
		params     *Params
		expectsErr bool
		expectsOut []string // Substrings to check for in output
	}{
		{
			name:       "default_output_bytes",
			params:     &Params{},
			expectsErr: false,
			expectsOut: []string{"Mem:", "Swap:", "KiB"},
		},
		{
			name:       "megabytes_output",
			params:     &Params{MegaBytes: true},
			expectsErr: false,
			expectsOut: []string{"Mem:", "Swap:", "MiB"},
		},
		{
			name:       "gigabytes_output",
			params:     &Params{GigaBytes: true},
			expectsErr: false,
			expectsOut: []string{"Mem:", "Swap:", "GiB"},
		},
		{
			name:       "both_mb_gb_should_prefer_gb",
			params:     &Params{MegaBytes: true, GigaBytes: true},
			expectsErr: false, // The cobra param enricher should handle this, but let's test what runFree does
			expectsOut: []string{"Mem:", "Swap:", "GiB"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdoutBuf bytes.Buffer
			// Temporarily redirect os.Stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runFree(tc.params)
			w.Close()
			os.Stdout = oldStdout // Restore original Stdout
			io.Copy(&stdoutBuf, r)

			if tc.expectsErr && err == nil {
				t.Errorf("Expected an error, but got none")
			} else if !tc.expectsErr && err != nil {
				t.Errorf("Did not expect an error, but got: %v", err)
			}

			output := stdoutBuf.String()
			for _, expectedSubstring := range tc.expectsOut {
				if !strings.Contains(output, expectedSubstring) {
					t.Errorf("Output \n%q\n does not contain expected substring %q", output, expectedSubstring)
				}
			}
		})
	}

	// Test scenario where gopsutil might fail (hard to mock without interface)
	// For now, relying on successful path due to gopsutil reliability.
}

// Helper to get dummy memory stats for testing, if we were to mock gopsutil
func getDummyVirtualMemoryStat() *mem.VirtualMemoryStat {
	return &mem.VirtualMemoryStat{
		Total:     16 * 1024 * 1024 * 1024, // 16GB
		Used:      8 * 1024 * 1024 * 1024,  // 8GB
		Free:      4 * 1024 * 1024 * 1024,  // 4GB
		Available: 6 * 1024 * 1024 * 1024,  // 6GB
		Shared:    1 * 1024 * 1024 * 1024,  // 1GB
		Buffers:   512 * 1024 * 1024,       // 0.5GB
		Cached:    1024 * 1024 * 1024,      // 1GB
	}
}

func getDummySwapMemoryStat() *mem.SwapMemoryStat {
	return &mem.SwapMemoryStat{
		Total: 4 * 1024 * 1024 * 1024, // 4GB
		Used:  2 * 1024 * 1024 * 1024, // 2GB
		Free:  2 * 1024 * 1024 * 1024, // 2GB
	}
}

// We cannot easily mock mem.VirtualMemory and mem.SwapMemory directly
// without changing the cmd_free.go implementation to accept an interface.
// For this test, we are relying on gopsutil to work correctly and
// checking output formatting, assuming it returns valid data.

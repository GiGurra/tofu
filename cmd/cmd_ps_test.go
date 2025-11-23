package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestPsCommand(t *testing.T) {
	// We can't easily predict PIDs or other users, so we'll test basic flags
	// and potentially the "current user" flag assuming the test runner has processes.

	tests := []struct {
		name        string
		params      *PsParams
		expectsErr  bool
		expectsOut  []string // Substrings to check for in output
		excludesOut []string // Substrings that should NOT appear
	}{
		{
			name:       "default_output",
			params:     &PsParams{},
			expectsErr: false,
			expectsOut: []string{"PID", "COMMAND"},
		},
		{
			name:       "full_output",
			params:     &PsParams{Full: true},
			expectsErr: false,
			expectsOut: []string{"PID", "PPID", "USER", "STATUS", "%CPU", "%MEM", "COMMAND"},
		},
		{
			name:       "current_user_filter",
			params:     &PsParams{Current: true},
			expectsErr: false,
			expectsOut: []string{"PID"}, // Should at least have a header and likely the test process
		},
		{
			name:       "filter_by_nonexistent_name",
			params:     &PsParams{Name: "this_process_name_should_not_exist_xyz123"},
			expectsErr: false,
			// Should output header but no rows (technically header is always printed)
			expectsOut: []string{"PID"},
		},
		{
			name:       "no_truncate_output",
			params:     &PsParams{Full: true, NoTruncate: true},
			expectsErr: false,
			expectsOut: []string{"PID", "COMMAND"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdoutBuf bytes.Buffer
			// Temporarily redirect os.Stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			wg := sync.WaitGroup{}
			wg.Go(func() {
				io.Copy(&stdoutBuf, r)
			})

			err := runPs(tc.params)

			w.Close()
			wg.Wait()
			os.Stdout = oldStdout // Restore original Stdout

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
			for _, excludedSubstring := range tc.excludesOut {
				if strings.Contains(output, excludedSubstring) {
					t.Errorf("Output \n%q\n SHOULD NOT contain substring %q", output, excludedSubstring)
				}
			}

			if tc.name == "filter_by_nonexistent_name" {
				// Check that it only has the header (1 line) or 2 lines if using tabwriter differently?
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 1 {
					// This check is a bit brittle if there actually IS a process named like that,
					// but highly unlikely.
					t.Logf("Warning: Found processes matching nonsense name: %v", lines[1:])
				}
			}
		})
	}
}

func TestPsFilterLogic(t *testing.T) {
	// This test intends to verify specific filtering logic with real system data
	// It's an integration test.

	myPid := int32(os.Getpid())

	t.Run("filter_by_self_pid", func(t *testing.T) {
		var stdoutBuf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		params := &PsParams{
			Pids: []int32{myPid},
		}
		err := runPs(params)
		w.Close()
		os.Stdout = oldStdout
		io.Copy(&stdoutBuf, r)

		if err != nil {
			t.Fatalf("runPs failed: %v", err)
		}

		output := stdoutBuf.String()
		if !strings.Contains(output, fmt.Sprintf("%d", myPid)) {
			t.Errorf("Output should contain own PID %d, got:\n%s", myPid, output)
		}
	})
}

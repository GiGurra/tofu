package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

// Helper to capture stdout
func captureOutput(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var captureErr error

	var buf bytes.Buffer
	wg := sync.WaitGroup{}
	wg.Go(func() {
		_, captureErr = io.Copy(&buf, r)
	})

	err := f()
	w.Close()

	wg.Wait()
	os.Stdout = old

	return buf.String(), errors.Join(err, captureErr)
}

func TestRunQr(t *testing.T) {
	// Simple test to ensure it runs and produces output containing ANSI codes
	params := &QrParams{
		Text: "test",
	}

	output, err := captureOutput(func() error {
		return runQr(params)
	})

	if err != nil {
		t.Fatalf("runQr failed: %v", err)
	}

	if output == "" {
		t.Error("Expected output, got empty string")
	}

	// Check for ANSI codes
	// Standard: Black (\033[40m) and White (\033[47m)
	if !strings.Contains(output, "\033[40m") {
		t.Error("Output should contain ANSI black background code")
	}
	if !strings.Contains(output, "\033[47m") {
		t.Error("Output should contain ANSI white background code")
	}
}

func TestRunQr_Invert(t *testing.T) {
	params := &QrParams{
		Text:   "test",
		Invert: true,
	}

	output, err := captureOutput(func() error {
		return runQr(params)
	})

	if err != nil {
		t.Fatalf("runQr failed: %v", err)
	}

	// Invert mode also uses the same codes, just swapped.
	// So we mostly check that it runs without error.
	if !strings.Contains(output, "\033[40m") {
		t.Error("Output should contain ANSI black background code")
	}
}

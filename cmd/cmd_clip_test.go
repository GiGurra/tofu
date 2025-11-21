package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunClip_CopyArgs(t *testing.T) {
	// Mock clipboard
	var capturedText string
	originalWrite := clipboardWriteAll
	clipboardWriteAll = func(text string) error {
		capturedText = text
		return nil
	}
	defer func() { clipboardWriteAll = originalWrite }()

	params := &ClipParams{}
	args := []string{"hello", "world"}
	var stdin strings.Reader
	var stdout bytes.Buffer

	err := runClip(params, args, &stdin, &stdout)
	if err != nil {
		t.Fatalf("runClip failed: %v", err)
	}

	if capturedText != "hello world" {
		t.Errorf("Expected 'hello world', got %q", capturedText)
	}
}

func TestRunClip_CopyStdin(t *testing.T) {
	// Mock clipboard
	var capturedText string
	originalWrite := clipboardWriteAll
	clipboardWriteAll = func(text string) error {
		capturedText = text
		return nil
	}
	defer func() { clipboardWriteAll = originalWrite }()

	params := &ClipParams{}
	input := "from stdin"
	stdin := strings.NewReader(input)
	var stdout bytes.Buffer

	err := runClip(params, []string{}, stdin, &stdout)
	if err != nil {
		t.Fatalf("runClip failed: %v", err)
	}

	if capturedText != "from stdin" {
		t.Errorf("Expected 'from stdin', got %q", capturedText)
	}
}

func TestRunClip_Paste(t *testing.T) {
	// Mock clipboard
	originalRead := clipboardReadAll
	clipboardReadAll = func() (string, error) {
		return "clipboard content", nil
	}
	defer func() { clipboardReadAll = originalRead }()

	params := &ClipParams{Paste: true}
	var stdin strings.Reader
	var stdout bytes.Buffer

	err := runClip(params, []string{}, &stdin, &stdout)
	if err != nil {
		t.Fatalf("runClip failed: %v", err)
	}

	if stdout.String() != "clipboard content" {
		t.Errorf("Expected 'clipboard content', got %q", stdout.String())
	}
}

func TestRunClip_PasteWithArgs(t *testing.T) {
	params := &ClipParams{Paste: true}
	var stdin strings.Reader
	var stdout bytes.Buffer

	err := runClip(params, []string{"arg"}, &stdin, &stdout)
	if err == nil {
		t.Fatal("Expected error when using arguments with --paste")
	}
}

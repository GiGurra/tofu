//go:build interop_age

package crypt

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// These tests require the 'age' CLI tool to be installed.
// Run with: go test -tags=interop_age ./cmd/crypt/...
//
// NOTE: age CLI requires a TTY for passphrase input.
// These tests use 'script' command to provide a pseudo-TTY on macOS/Linux.
// On systems without 'script', tests will be skipped.
//
// Manual testing:
//   # tofu encrypt -> age decrypt
//   echo "test" > /tmp/test.txt
//   tofu crypt encrypt -p secret /tmp/test.txt
//   age -d /tmp/test.txt.age  # enter "secret" when prompted
//
//   # age encrypt -> tofu decrypt
//   age -p -o /tmp/test2.age /tmp/test.txt  # enter password twice
//   tofu crypt decrypt -p <password> /tmp/test2.age

func hasScriptCommand() bool {
	_, err := exec.LookPath("script")
	return err == nil
}

func runAgeWithPassphrase(args []string, password string) ([]byte, error) {
	// Use 'script' to provide a PTY for age
	// On macOS: script -q /dev/null age ...
	// On Linux: script -q -c "age ..." /dev/null

	ageCmd := append([]string{"age"}, args...)

	// Try macOS style first
	cmd := exec.Command("script", append([]string{"-q", "/dev/null"}, ageCmd...)...)
	cmd.Stdin = bytes.NewBufferString(password + "\n" + password + "\n")
	return cmd.CombinedOutput()
}

func TestInteropTofuEncryptAgeDecrypt(t *testing.T) {
	if _, err := exec.LookPath("age"); err != nil {
		t.Skip("age CLI not installed, skipping interop test")
	}
	if !hasScriptCommand() {
		t.Skip("'script' command not available for PTY simulation, skipping interop test")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	encFile := filepath.Join(tmpDir, "input.txt.age")
	decFile := filepath.Join(tmpDir, "decrypted.txt")

	content := []byte("Hello from tofu! Testing interoperability with age CLI.")
	password := "testpassword123"

	// Write input file
	if err := os.WriteFile(inputFile, content, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Encrypt with tofu
	if err := encryptFileAge(inputFile, encFile, password); err != nil {
		t.Fatalf("tofu encryption failed: %v", err)
	}

	// Decrypt with age CLI using script for PTY
	output, err := runAgeWithPassphrase([]string{"-d", "-o", decFile, encFile}, password)
	if err != nil {
		t.Fatalf("age decryption failed: %v\nOutput: %s", err, output)
	}

	// Verify content
	decContent, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	if !bytes.Equal(decContent, content) {
		t.Errorf("content mismatch: got %q, want %q", decContent, content)
	}
}

func TestInteropAgeEncryptTofuDecrypt(t *testing.T) {
	if _, err := exec.LookPath("age"); err != nil {
		t.Skip("age CLI not installed, skipping interop test")
	}
	if !hasScriptCommand() {
		t.Skip("'script' command not available for PTY simulation, skipping interop test")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	encFile := filepath.Join(tmpDir, "input.txt.age")
	decFile := filepath.Join(tmpDir, "decrypted.txt")

	content := []byte("Hello from age CLI! Testing interoperability with tofu.")
	password := "testpassword123"

	// Write input file
	if err := os.WriteFile(inputFile, content, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Encrypt with age CLI
	output, err := runAgeWithPassphrase([]string{"-p", "-o", encFile, inputFile}, password)
	if err != nil {
		t.Fatalf("age encryption failed: %v\nOutput: %s", err, output)
	}

	// Decrypt with tofu
	if err := decryptFileAge(encFile, decFile, password); err != nil {
		t.Fatalf("tofu decryption failed: %v", err)
	}

	// Verify content
	decContent, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	if !bytes.Equal(decContent, content) {
		t.Errorf("content mismatch: got %q, want %q", decContent, content)
	}
}

func TestInteropAgeBinaryData(t *testing.T) {
	if _, err := exec.LookPath("age"); err != nil {
		t.Skip("age CLI not installed, skipping interop test")
	}
	if !hasScriptCommand() {
		t.Skip("'script' command not available for PTY simulation, skipping interop test")
	}

	tmpDir := t.TempDir()
	password := "testpassword123"

	// Binary content with null bytes and all byte values
	content := make([]byte, 256)
	for i := range content {
		content[i] = byte(i)
	}

	t.Run("tofu_encrypt_age_decrypt", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "binary1.bin")
		encFile := filepath.Join(tmpDir, "binary1.bin.age")
		decFile := filepath.Join(tmpDir, "binary1_dec.bin")

		os.WriteFile(inputFile, content, 0644)

		if err := encryptFileAge(inputFile, encFile, password); err != nil {
			t.Fatalf("tofu encryption failed: %v", err)
		}

		output, err := runAgeWithPassphrase([]string{"-d", "-o", decFile, encFile}, password)
		if err != nil {
			t.Fatalf("age decryption failed: %v\nOutput: %s", err, output)
		}

		decContent, _ := os.ReadFile(decFile)
		if !bytes.Equal(decContent, content) {
			t.Error("binary content mismatch")
		}
	})

	t.Run("age_encrypt_tofu_decrypt", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "binary2.bin")
		encFile := filepath.Join(tmpDir, "binary2.bin.age")
		decFile := filepath.Join(tmpDir, "binary2_dec.bin")

		os.WriteFile(inputFile, content, 0644)

		output, err := runAgeWithPassphrase([]string{"-p", "-o", encFile, inputFile}, password)
		if err != nil {
			t.Fatalf("age encryption failed: %v\nOutput: %s", err, output)
		}

		if err := decryptFileAge(encFile, decFile, password); err != nil {
			t.Fatalf("tofu decryption failed: %v", err)
		}

		decContent, _ := os.ReadFile(decFile)
		if !bytes.Equal(decContent, content) {
			t.Error("binary content mismatch")
		}
	})
}

//go:build interop_openssl

package crypt

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// These tests require the 'openssl' CLI tool to be installed.
// Run with: go test -tags=interop_openssl ./cmd/crypt/...
//
// Note: tofu uses PBKDF2 with 600000 iterations for openssl compatibility.
// When decrypting with openssl CLI, you must specify: -iter 600000

func TestInteropTofuEncryptOpenSSLDecrypt(t *testing.T) {
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl CLI not installed, skipping interop test")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	encFile := filepath.Join(tmpDir, "input.txt.enc")
	decFile := filepath.Join(tmpDir, "decrypted.txt")

	content := []byte("Hello from tofu! Testing interoperability with openssl CLI.")
	password := "testpassword123"

	// Write input file
	if err := os.WriteFile(inputFile, content, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Encrypt with tofu
	if err := encryptFileOpenSSL(inputFile, encFile, password); err != nil {
		t.Fatalf("tofu encryption failed: %v", err)
	}

	// Decrypt with openssl CLI
	// Note: must specify -iter to match our iteration count
	cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-pbkdf2",
		"-iter", "600000",
		"-in", encFile, "-out", decFile,
		"-pass", "pass:"+password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("openssl decryption failed: %v\nOutput: %s", err, output)
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

func TestInteropOpenSSLEncryptTofuDecrypt(t *testing.T) {
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl CLI not installed, skipping interop test")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	encFile := filepath.Join(tmpDir, "input.txt.enc")
	decFile := filepath.Join(tmpDir, "decrypted.txt")

	content := []byte("Hello from openssl CLI! Testing interoperability with tofu.")
	password := "testpassword123"

	// Write input file
	if err := os.WriteFile(inputFile, content, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Encrypt with openssl CLI
	// Note: must specify -iter to match our iteration count
	cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-pbkdf2",
		"-iter", "600000",
		"-in", inputFile, "-out", encFile,
		"-pass", "pass:"+password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("openssl encryption failed: %v\nOutput: %s", err, output)
	}

	// Decrypt with tofu
	if err := decryptFileOpenSSL(encFile, decFile, password); err != nil {
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

func TestInteropOpenSSLTofuRoundtripLargeFile(t *testing.T) {
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl CLI not installed, skipping interop test")
	}

	tmpDir := t.TempDir()
	password := "testpassword123"

	// Test with 1MB file
	content := bytes.Repeat([]byte("Large file content for interop test. "), 30000)

	t.Run("tofu->openssl->tofu", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "large1.txt")
		encFile := filepath.Join(tmpDir, "large1.txt.enc")
		midFile := filepath.Join(tmpDir, "large1_mid.txt")
		reencFile := filepath.Join(tmpDir, "large1_mid.txt.enc")
		finalFile := filepath.Join(tmpDir, "large1_final.txt")

		os.WriteFile(inputFile, content, 0644)

		// tofu encrypt
		if err := encryptFileOpenSSL(inputFile, encFile, password); err != nil {
			t.Fatalf("tofu encryption failed: %v", err)
		}

		// openssl decrypt
		cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-pbkdf2",
			"-iter", "600000",
			"-in", encFile, "-out", midFile,
			"-pass", "pass:"+password)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("openssl decryption failed: %v\nOutput: %s", err, output)
		}

		// openssl encrypt
		cmd = exec.Command("openssl", "enc", "-aes-256-cbc", "-pbkdf2",
			"-iter", "600000",
			"-in", midFile, "-out", reencFile,
			"-pass", "pass:"+password)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("openssl encryption failed: %v\nOutput: %s", err, output)
		}

		// tofu decrypt
		if err := decryptFileOpenSSL(reencFile, finalFile, password); err != nil {
			t.Fatalf("tofu decryption failed: %v", err)
		}

		// Verify
		finalContent, _ := os.ReadFile(finalFile)
		if !bytes.Equal(finalContent, content) {
			t.Error("content mismatch after roundtrip")
		}
	})
}

func TestInteropOpenSSLBinaryData(t *testing.T) {
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl CLI not installed, skipping interop test")
	}

	tmpDir := t.TempDir()
	password := "testpassword123"

	// Binary content with null bytes and all byte values
	content := make([]byte, 256)
	for i := range content {
		content[i] = byte(i)
	}

	t.Run("tofu_encrypt_openssl_decrypt", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "binary1.bin")
		encFile := filepath.Join(tmpDir, "binary1.bin.enc")
		decFile := filepath.Join(tmpDir, "binary1_dec.bin")

		os.WriteFile(inputFile, content, 0644)

		if err := encryptFileOpenSSL(inputFile, encFile, password); err != nil {
			t.Fatalf("tofu encryption failed: %v", err)
		}

		cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-pbkdf2",
			"-iter", "600000",
			"-in", encFile, "-out", decFile,
			"-pass", "pass:"+password)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("openssl decryption failed: %v\nOutput: %s", err, output)
		}

		decContent, _ := os.ReadFile(decFile)
		if !bytes.Equal(decContent, content) {
			t.Error("binary content mismatch")
		}
	})

	t.Run("openssl_encrypt_tofu_decrypt", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "binary2.bin")
		encFile := filepath.Join(tmpDir, "binary2.bin.enc")
		decFile := filepath.Join(tmpDir, "binary2_dec.bin")

		os.WriteFile(inputFile, content, 0644)

		cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-pbkdf2",
			"-iter", "600000",
			"-in", inputFile, "-out", encFile,
			"-pass", "pass:"+password)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("openssl encryption failed: %v\nOutput: %s", err, output)
		}

		if err := decryptFileOpenSSL(encFile, decFile, password); err != nil {
			t.Fatalf("tofu decryption failed: %v", err)
		}

		decContent, _ := os.ReadFile(decFile)
		if !bytes.Equal(decContent, content) {
			t.Error("binary content mismatch")
		}
	})
}

func TestInteropOpenSSLEmptyFile(t *testing.T) {
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl CLI not installed, skipping interop test")
	}

	tmpDir := t.TempDir()
	password := "testpassword123"
	content := []byte{}

	t.Run("tofu_encrypt_openssl_decrypt", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "empty1.txt")
		encFile := filepath.Join(tmpDir, "empty1.txt.enc")
		decFile := filepath.Join(tmpDir, "empty1_dec.txt")

		os.WriteFile(inputFile, content, 0644)

		if err := encryptFileOpenSSL(inputFile, encFile, password); err != nil {
			t.Fatalf("tofu encryption failed: %v", err)
		}

		cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-pbkdf2",
			"-iter", "600000",
			"-in", encFile, "-out", decFile,
			"-pass", "pass:"+password)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("openssl decryption failed: %v\nOutput: %s", err, output)
		}

		decContent, _ := os.ReadFile(decFile)
		if !bytes.Equal(decContent, content) {
			t.Errorf("content mismatch: got %d bytes, want %d bytes", len(decContent), len(content))
		}
	})

	t.Run("openssl_encrypt_tofu_decrypt", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "empty2.txt")
		encFile := filepath.Join(tmpDir, "empty2.txt.enc")
		decFile := filepath.Join(tmpDir, "empty2_dec.txt")

		os.WriteFile(inputFile, content, 0644)

		cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-pbkdf2",
			"-iter", "600000",
			"-in", inputFile, "-out", encFile,
			"-pass", "pass:"+password)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("openssl encryption failed: %v\nOutput: %s", err, output)
		}

		if err := decryptFileOpenSSL(encFile, decFile, password); err != nil {
			t.Fatalf("tofu decryption failed: %v", err)
		}

		decContent, _ := os.ReadFile(decFile)
		if !bytes.Equal(decContent, content) {
			t.Errorf("content mismatch: got %d bytes, want %d bytes", len(decContent), len(content))
		}
	})
}

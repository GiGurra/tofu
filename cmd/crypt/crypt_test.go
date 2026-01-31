package crypt

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundtripAge(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{"small", []byte("Hello, World!")},
		{"empty", []byte{}},
		{"large", bytes.Repeat([]byte("A"), 1024*1024)}, // 1MB
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界! 🎉 Ñoño αβγδ Привет")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "input.txt")
			encFile := filepath.Join(tmpDir, "input.txt.age")
			decFile := filepath.Join(tmpDir, "decrypted.txt")

			// Write input file
			if err := os.WriteFile(inputFile, tt.content, 0644); err != nil {
				t.Fatalf("failed to write input file: %v", err)
			}

			password := "testpassword123"

			// Encrypt
			if err := encryptFileAge(inputFile, encFile, password); err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			// Verify encrypted file exists and is different
			encContent, err := os.ReadFile(encFile)
			if err != nil {
				t.Fatalf("failed to read encrypted file: %v", err)
			}

			if len(tt.content) > 0 && bytes.Equal(encContent, tt.content) {
				t.Error("encrypted content should differ from original")
			}

			// Verify age format header
			if !strings.HasPrefix(string(encContent), "age-encryption.org/") {
				t.Error("encrypted file should start with age header")
			}

			// Decrypt
			if err := decryptFileAge(encFile, decFile, password); err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			// Verify decrypted content matches original
			decContent, err := os.ReadFile(decFile)
			if err != nil {
				t.Fatalf("failed to read decrypted file: %v", err)
			}

			if !bytes.Equal(decContent, tt.content) {
				t.Errorf("decrypted content doesn't match original: got %d bytes, want %d bytes", len(decContent), len(tt.content))
			}
		})
	}
}

func TestEncryptDecryptRoundtripOpenSSL(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{"small", []byte("Hello, World!")},
		{"empty", []byte{}},
		{"large", bytes.Repeat([]byte("B"), 1024*1024)}, // 1MB
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界! 🎉 Ñoño αβγδ Привет")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "input.txt")
			encFile := filepath.Join(tmpDir, "input.txt.enc")
			decFile := filepath.Join(tmpDir, "decrypted.txt")

			// Write input file
			if err := os.WriteFile(inputFile, tt.content, 0644); err != nil {
				t.Fatalf("failed to write input file: %v", err)
			}

			password := "testpassword123"

			// Encrypt
			if err := encryptFileOpenSSL(inputFile, encFile, password); err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			// Verify encrypted file exists
			encContent, err := os.ReadFile(encFile)
			if err != nil {
				t.Fatalf("failed to read encrypted file: %v", err)
			}

			// Verify OpenSSL format header
			if !strings.HasPrefix(string(encContent), opensslSaltHeader) {
				t.Error("encrypted file should start with Salted__ header")
			}

			// Decrypt
			if err := decryptFileOpenSSL(encFile, decFile, password); err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			// Verify decrypted content matches original
			decContent, err := os.ReadFile(decFile)
			if err != nil {
				t.Fatalf("failed to read decrypted file: %v", err)
			}

			if !bytes.Equal(decContent, tt.content) {
				t.Errorf("decrypted content doesn't match original: got %d bytes, want %d bytes", len(decContent), len(tt.content))
			}
		})
	}
}

func TestDecryptWrongPasswordAge(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "secret.txt")
	encFile := filepath.Join(tmpDir, "secret.txt.age")
	decFile := filepath.Join(tmpDir, "decrypted.txt")

	content := []byte("This is a secret message")
	if err := os.WriteFile(inputFile, content, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Encrypt with one password
	if err := encryptFileAge(inputFile, encFile, "correctpassword"); err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with wrong password
	err := decryptFileAge(encFile, decFile, "wrongpassword")
	if err == nil {
		t.Error("decryption should fail with wrong password")
	}
}

func TestDecryptWrongPasswordOpenSSL(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "secret.txt")
	encFile := filepath.Join(tmpDir, "secret.txt.enc")
	decFile := filepath.Join(tmpDir, "decrypted.txt")

	content := []byte("This is a secret message")
	if err := os.WriteFile(inputFile, content, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Encrypt with one password
	if err := encryptFileOpenSSL(inputFile, encFile, "correctpassword"); err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with wrong password
	err := decryptFileOpenSSL(encFile, decFile, "wrongpassword")
	if err == nil {
		t.Error("decryption should fail with wrong password")
	}
}

func TestDetectFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create age format file
	ageFile := filepath.Join(tmpDir, "test.age")
	ageContent := []byte("age-encryption.org/v1\n-> scrypt abc123 18\nrest of file")
	os.WriteFile(ageFile, ageContent, 0644)

	// Create OpenSSL format file
	opensslFile := filepath.Join(tmpDir, "test.enc")
	opensslContent := append([]byte("Salted__12345678"), []byte("encrypted data")...)
	os.WriteFile(opensslFile, opensslContent, 0644)

	// Create unknown format file
	unknownFile := filepath.Join(tmpDir, "test.bin")
	os.WriteFile(unknownFile, []byte("random data"), 0644)

	tests := []struct {
		file string
		want string
	}{
		{ageFile, "age"},
		{opensslFile, "openssl"},
		{unknownFile, "age"}, // defaults to age
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.file), func(t *testing.T) {
			got, err := detectFormat(tt.file)
			if err != nil {
				t.Fatalf("detectFormat failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("detectFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetermineDecryptOutputPath(t *testing.T) {
	tests := []struct {
		input  string
		format string
		want   string
	}{
		{"file.txt.age", "age", "file.txt"},
		{"file.txt.enc", "openssl", "file.txt"},
		{"file.age", "age", "file"},
		{"file.enc", "openssl", "file"},
		{"file.txt", "age", "file.txt.dec"},
		{"file", "age", "file.dec"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := determineDecryptOutputPath(tt.input, tt.format)
			if got != tt.want {
				t.Errorf("determineDecryptOutputPath(%q, %q) = %q, want %q", tt.input, tt.format, got, tt.want)
			}
		})
	}
}

func TestRunEncryptValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Test no files specified
	err := runEncrypt(&EncryptParams{
		Password: "test",
		Format:   "age",
	})
	if err == nil || err.Error() != "no files specified" {
		t.Errorf("expected 'no files specified' error, got: %v", err)
	}

	// Test -o with multiple files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	os.WriteFile(file1, []byte("test1"), 0644)
	os.WriteFile(file2, []byte("test2"), 0644)

	err = runEncrypt(&EncryptParams{
		Files:    []string{file1, file2},
		Output:   "single.enc",
		Password: "test",
		Format:   "age",
	})
	if err == nil || err.Error() != "-o can only be used with a single input file" {
		t.Errorf("expected '-o can only be used with a single input file' error, got: %v", err)
	}

	// Test unknown format
	err = runEncrypt(&EncryptParams{
		Files:    []string{file1},
		Password: "test",
		Format:   "unknown",
	})
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected 'unknown format' error, got: %v", err)
	}
}

func TestRunDecryptValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Test no files specified
	err := runDecrypt(&DecryptParams{
		Password: "test",
		Format:   "auto",
	})
	if err == nil || err.Error() != "no files specified" {
		t.Errorf("expected 'no files specified' error, got: %v", err)
	}

	// Test -o with multiple files
	file1 := filepath.Join(tmpDir, "file1.age")
	file2 := filepath.Join(tmpDir, "file2.age")
	os.WriteFile(file1, []byte("test1"), 0644)
	os.WriteFile(file2, []byte("test2"), 0644)

	err = runDecrypt(&DecryptParams{
		Files:    []string{file1, file2},
		Output:   "single.txt",
		Password: "test",
		Format:   "auto",
	})
	if err == nil || err.Error() != "-o can only be used with a single input file" {
		t.Errorf("expected '-o can only be used with a single input file' error, got: %v", err)
	}
}

func TestOutputFileNamingAge(t *testing.T) {
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "document.pdf")
	os.WriteFile(inputFile, []byte("pdf content"), 0644)

	err := runEncrypt(&EncryptParams{
		Files:    []string{inputFile},
		Password: "test",
		Format:   "age",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	encFile := inputFile + ".age"
	if _, err := os.Stat(encFile); os.IsNotExist(err) {
		t.Error("encrypted file should be named with .age extension")
	}

	// Test decrypt removes .age
	os.Remove(inputFile)

	err = runDecrypt(&DecryptParams{
		Files:    []string{encFile},
		Password: "test",
		Format:   "auto",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Error("decrypted file should have .age extension removed")
	}
}

func TestOutputFileNamingOpenSSL(t *testing.T) {
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "document.pdf")
	os.WriteFile(inputFile, []byte("pdf content"), 0644)

	err := runEncrypt(&EncryptParams{
		Files:    []string{inputFile},
		Password: "test",
		Format:   "openssl",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	encFile := inputFile + ".enc"
	if _, err := os.Stat(encFile); os.IsNotExist(err) {
		t.Error("encrypted file should be named with .enc extension")
	}

	// Test decrypt removes .enc
	os.Remove(inputFile)

	err = runDecrypt(&DecryptParams{
		Files:    []string{encFile},
		Password: "test",
		Format:   "auto",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Error("decrypted file should have .enc extension removed")
	}
}

func TestForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	encFile := filepath.Join(tmpDir, "input.txt.age")

	os.WriteFile(inputFile, []byte("original"), 0644)
	os.WriteFile(encFile, []byte("existing"), 0644)

	// Without force, should fail
	err := runEncrypt(&EncryptParams{
		Files:    []string{inputFile},
		Password: "test",
		Format:   "age",
		Keep:     true,
		Force:    false,
	})
	if err == nil {
		t.Error("encryption should fail when output exists without -F")
	}

	// With force, should succeed
	err = runEncrypt(&EncryptParams{
		Files:    []string{inputFile},
		Password: "test",
		Format:   "age",
		Keep:     true,
		Force:    true,
	})
	if err != nil {
		t.Errorf("encryption should succeed with -F: %v", err)
	}
}

func TestKeepOriginalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	encFile := filepath.Join(tmpDir, "input.txt.age")

	os.WriteFile(inputFile, []byte("test content"), 0644)

	// With keep=true, original should remain
	err := runEncrypt(&EncryptParams{
		Files:    []string{inputFile},
		Password: "test",
		Format:   "age",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Error("original file should exist with -k flag")
	}

	// Without keep, original should be removed
	inputFile2 := filepath.Join(tmpDir, "input2.txt")
	os.WriteFile(inputFile2, []byte("test content 2"), 0644)

	err = runEncrypt(&EncryptParams{
		Files:    []string{inputFile2},
		Password: "test",
		Format:   "age",
		Keep:     false,
	})
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if _, err := os.Stat(inputFile2); !os.IsNotExist(err) {
		t.Error("original file should be removed without -k flag")
	}

	// For decrypt
	err = runDecrypt(&DecryptParams{
		Files:    []string{encFile},
		Password: "test",
		Format:   "auto",
		Keep:     false,
		Force:    true,
	})
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if _, err := os.Stat(encFile); !os.IsNotExist(err) {
		t.Error("encrypted file should be removed without -k flag")
	}
}

func TestEncryptNonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistent := filepath.Join(tmpDir, "doesnotexist.txt")

	err := runEncrypt(&EncryptParams{
		Files:    []string{nonexistent},
		Password: "test",
		Format:   "age",
	})
	if err == nil {
		t.Error("encryption should fail for nonexistent file")
	}
}

func TestMultipleFilesEncryption(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{
		filepath.Join(tmpDir, "file1.txt"),
		filepath.Join(tmpDir, "file2.txt"),
		filepath.Join(tmpDir, "file3.txt"),
	}

	for i, f := range files {
		os.WriteFile(f, []byte("content "+string(rune('A'+i))), 0644)
	}

	err := runEncrypt(&EncryptParams{
		Files:    files,
		Password: "test",
		Format:   "age",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Verify all encrypted files exist
	encFiles := make([]string, len(files))
	for i, f := range files {
		encFiles[i] = f + ".age"
		if _, err := os.Stat(encFiles[i]); os.IsNotExist(err) {
			t.Errorf("encrypted file should exist: %s", encFiles[i])
		}
	}

	// Remove originals
	for _, f := range files {
		os.Remove(f)
	}

	// Decrypt all
	err = runDecrypt(&DecryptParams{
		Files:    encFiles,
		Password: "test",
		Format:   "auto",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Verify all decrypted files exist and have correct content
	for i, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("failed to read decrypted file %s: %v", f, err)
			continue
		}
		expected := "content " + string(rune('A'+i))
		if string(content) != expected {
			t.Errorf("decrypted content mismatch for %s: got %q, want %q", f, content, expected)
		}
	}
}

func TestEncryptPreservesFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	encFile := filepath.Join(tmpDir, "input.txt.age")

	// Create file with specific permissions
	if err := os.WriteFile(inputFile, []byte("test"), 0600); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Encrypt
	if err := encryptFileAge(inputFile, encFile, "password"); err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Check permissions
	inputInfo, _ := os.Stat(inputFile)
	encInfo, _ := os.Stat(encFile)

	if inputInfo.Mode() != encInfo.Mode() {
		t.Errorf("encrypted file permissions = %v, want %v", encInfo.Mode(), inputInfo.Mode())
	}
}

func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		blockSize int
	}{
		{"empty", []byte{}, 16},
		{"one byte", []byte{0x01}, 16},
		{"exactly block size", bytes.Repeat([]byte{0x42}, 16), 16},
		{"block size + 1", bytes.Repeat([]byte{0x42}, 17), 16},
		{"multiple blocks", bytes.Repeat([]byte{0x42}, 100), 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := pkcs7Pad(tt.input, tt.blockSize)

			// Verify padded length is multiple of block size
			if len(padded)%tt.blockSize != 0 {
				t.Errorf("padded length %d is not multiple of %d", len(padded), tt.blockSize)
			}

			// Verify padding is at least 1 byte
			if len(padded) < len(tt.input)+1 {
				t.Error("padding should add at least 1 byte")
			}

			// Unpad and verify
			unpadded, err := pkcs7Unpad(padded)
			if err != nil {
				t.Fatalf("unpad failed: %v", err)
			}

			if !bytes.Equal(unpadded, tt.input) {
				t.Errorf("unpadded doesn't match original: got %v, want %v", unpadded, tt.input)
			}
		})
	}
}

func TestPKCS7UnpadInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"zero padding", []byte{0x41, 0x41, 0x00}},
		{"padding too large", []byte{0x41, 0x41, 0x20}},                // 32 > 16
		{"inconsistent padding", []byte{0x41, 0x41, 0x01, 0x03, 0x03}}, // last byte says 3, but data[2]=0x01 not 0x03
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pkcs7Unpad(tt.input)
			if err == nil {
				t.Error("expected error for invalid padding")
			}
		})
	}
}

func TestDecryptCorruptedOpenSSL(t *testing.T) {
	tmpDir := t.TempDir()
	encFile := filepath.Join(tmpDir, "corrupted.enc")
	decFile := filepath.Join(tmpDir, "decrypted.txt")

	tests := []struct {
		name    string
		content []byte
	}{
		{"too short", []byte("short")},
		{"wrong header", []byte("NotSalt_12345678ciphertext")},
		{"invalid ciphertext length", []byte("Salted__12345678abc")}, // not multiple of 16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(encFile, tt.content, 0644); err != nil {
				t.Fatalf("failed to write corrupted file: %v", err)
			}

			err := decryptFileOpenSSL(encFile, decFile, "password")
			if err == nil {
				t.Error("decryption should fail for corrupted file")
			}
		})
	}
}

func TestAutoFormatDetectionIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("test content for auto detection")

	// Test age format
	ageInput := filepath.Join(tmpDir, "age_input.txt")
	ageEnc := filepath.Join(tmpDir, "age_input.txt.age")
	ageDec := filepath.Join(tmpDir, "age_dec.txt")
	os.WriteFile(ageInput, content, 0644)

	err := runEncrypt(&EncryptParams{
		Files:    []string{ageInput},
		Password: "test",
		Format:   "age",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("age encryption failed: %v", err)
	}

	// Decrypt with auto detection
	err = runDecrypt(&DecryptParams{
		Files:    []string{ageEnc},
		Output:   ageDec,
		Password: "test",
		Format:   "auto",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("age decryption with auto detection failed: %v", err)
	}

	decContent, _ := os.ReadFile(ageDec)
	if !bytes.Equal(decContent, content) {
		t.Error("age auto-detect decryption content mismatch")
	}

	// Test openssl format
	opensslInput := filepath.Join(tmpDir, "openssl_input.txt")
	opensslEnc := filepath.Join(tmpDir, "openssl_input.txt.enc")
	opensslDec := filepath.Join(tmpDir, "openssl_dec.txt")
	os.WriteFile(opensslInput, content, 0644)

	err = runEncrypt(&EncryptParams{
		Files:    []string{opensslInput},
		Password: "test",
		Format:   "openssl",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("openssl encryption failed: %v", err)
	}

	// Decrypt with auto detection
	err = runDecrypt(&DecryptParams{
		Files:    []string{opensslEnc},
		Output:   opensslDec,
		Password: "test",
		Format:   "auto",
		Keep:     true,
	})
	if err != nil {
		t.Fatalf("openssl decryption with auto detection failed: %v", err)
	}

	decContent, _ = os.ReadFile(opensslDec)
	if !bytes.Equal(decContent, content) {
		t.Error("openssl auto-detect decryption content mismatch")
	}
}

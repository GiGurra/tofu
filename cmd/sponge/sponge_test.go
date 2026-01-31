package sponge

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_SmallInput(t *testing.T) {
	input := "Hello, World!"
	var output bytes.Buffer

	err := Run(strings.NewReader(input), &output, 1024)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Expected %q, got %q", input, output.String())
	}
}

func TestRun_EmptyInput(t *testing.T) {
	input := ""
	var output bytes.Buffer

	err := Run(strings.NewReader(input), &output, 1024)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

func TestRun_MultilineInput(t *testing.T) {
	input := "Line 1\nLine 2\nLine 3\n"
	var output bytes.Buffer

	err := Run(strings.NewReader(input), &output, 1024)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Expected %q, got %q", input, output.String())
	}
}

func TestRun_ExactlyMaxSize(t *testing.T) {
	input := "12345" // 5 bytes
	var output bytes.Buffer

	err := Run(strings.NewReader(input), &output, 5)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Expected %q, got %q", input, output.String())
	}
}

func TestRun_ExceedsMaxSize_UseTempFile(t *testing.T) {
	// Create input larger than maxSize to force temp file usage
	input := strings.Repeat("x", 1000)
	var output bytes.Buffer

	// maxSize of 100 should trigger temp file buffering
	err := Run(strings.NewReader(input), &output, 100)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Expected length %d, got length %d", len(input), len(output.String()))
	}
}

func TestRun_LargeInput_TempFile(t *testing.T) {
	// Generate 100KB of random data
	inputSize := 100 * 1024
	inputData := make([]byte, inputSize)
	if _, err := rand.Read(inputData); err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	var output bytes.Buffer

	// maxSize of 10KB should trigger temp file usage
	err := Run(bytes.NewReader(inputData), &output, 10*1024)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !bytes.Equal(output.Bytes(), inputData) {
		t.Errorf("Output does not match input. Input length: %d, Output length: %d", len(inputData), output.Len())
	}
}

func TestRun_BinaryData(t *testing.T) {
	// Test with binary data including null bytes
	inputData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0x00, 0x10}
	var output bytes.Buffer

	err := Run(bytes.NewReader(inputData), &output, 1024)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !bytes.Equal(output.Bytes(), inputData) {
		t.Errorf("Binary data mismatch")
	}
}

func TestRun_BinaryData_TempFile(t *testing.T) {
	// Test binary data with temp file path
	inputData := make([]byte, 1000)
	for i := range inputData {
		inputData[i] = byte(i % 256)
	}
	var output bytes.Buffer

	// Force temp file usage with small maxSize
	err := Run(bytes.NewReader(inputData), &output, 100)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !bytes.Equal(output.Bytes(), inputData) {
		t.Errorf("Binary data mismatch with temp file")
	}
}

func TestRun_ZeroMaxSize(t *testing.T) {
	// maxSize of 0 should immediately use temp file for any input
	input := "test"
	var output bytes.Buffer

	err := Run(strings.NewReader(input), &output, 0)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Expected %q, got %q", input, output.String())
	}
}

// errorReader is a reader that returns an error after a certain number of reads
type errorReader struct {
	readCount    int
	errorOnRead  int
	readErr      error
	bytesPerRead int
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	r.readCount++
	if r.readCount >= r.errorOnRead {
		return 0, r.readErr
	}
	// Return some data
	n = r.bytesPerRead
	if n > len(p) {
		n = len(p)
	}
	for i := 0; i < n; i++ {
		p[i] = 'x'
	}
	return n, nil
}

func TestRun_ReadError(t *testing.T) {
	reader := &errorReader{
		errorOnRead:  3,
		readErr:      io.ErrUnexpectedEOF,
		bytesPerRead: 100,
	}
	var output bytes.Buffer

	err := Run(reader, &output, 1024)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// Buffer tests

func TestBuffer_MemoryOnly(t *testing.T) {
	buf := NewBuffer(1024)
	defer buf.Close()

	input := "Hello, World!"
	_, err := buf.ReadFrom(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadFrom error: %v", err)
	}

	if buf.UsingTempFile() {
		t.Error("Expected memory buffer, got temp file")
	}
	if buf.Size() != int64(len(input)) {
		t.Errorf("Expected size %d, got %d", len(input), buf.Size())
	}

	var output bytes.Buffer
	_, err = buf.WriteTo(&output)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Expected %q, got %q", input, output.String())
	}
}

func TestBuffer_SpillsToTempFile(t *testing.T) {
	buf := NewBuffer(100)
	defer buf.Close()

	input := strings.Repeat("x", 1000)
	_, err := buf.ReadFrom(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadFrom error: %v", err)
	}

	if !buf.UsingTempFile() {
		t.Error("Expected temp file, got memory buffer")
	}
	if buf.Size() != int64(len(input)) {
		t.Errorf("Expected size %d, got %d", len(input), buf.Size())
	}

	var output bytes.Buffer
	_, err = buf.WriteTo(&output)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}

	if output.String() != input {
		t.Errorf("Output length mismatch: expected %d, got %d", len(input), output.Len())
	}
}

// Same-file read/write tests - the classic sponge use case

// TestRunToFile_SameFile_Memory tests reading from a file and writing back to
// the same file using memory buffering. This is the classic sponge use case.
func TestRunToFile_SameFile_Memory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create initial file with content
	originalContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open file for reading
	inputFile, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	// Use RunToFile to read from file handle and write back to same path
	// This simulates: cat file | sponge file
	err = RunToFile(inputFile, testFile, 1024*1024)
	inputFile.Close()
	if err != nil {
		t.Fatalf("RunToFile error: %v", err)
	}

	// Verify the file content is unchanged
	resultContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	if string(resultContent) != originalContent {
		t.Errorf("File content changed!\nExpected: %q\nGot: %q", originalContent, string(resultContent))
	}
}

// TestRunToFile_SameFile_TempFile tests reading from a file and writing back
// to the same file, with content large enough to trigger temp file buffering.
func TestRunToFile_SameFile_TempFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create a large file (100KB of random data)
	originalContent := make([]byte, 100*1024)
	if _, err := rand.Read(originalContent); err != nil {
		t.Fatalf("Failed to generate random content: %v", err)
	}
	if err := os.WriteFile(testFile, originalContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open file for reading
	inputFile, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	// Use RunToFile with small maxSize to force temp file usage
	err = RunToFile(inputFile, testFile, 10*1024)
	inputFile.Close()
	if err != nil {
		t.Fatalf("RunToFile error: %v", err)
	}

	// Verify the file content is unchanged
	resultContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	if !bytes.Equal(resultContent, originalContent) {
		t.Errorf("File content changed! Original length: %d, Result length: %d",
			len(originalContent), len(resultContent))
	}
}

// TestRunToFile_TransformPipeline simulates a realistic pipeline where content
// is transformed before being written back to the same file.
// Simulates: cat file | grep -v pattern | sponge file
func TestRunToFile_TransformPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "pipeline.txt")

	// Create initial file
	originalContent := "apple\nbanana\ncherry\ndate\nelderberry\n"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open file for reading
	inputFile, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	// Create a filtering reader that removes "banana" line
	// This simulates: cat file | grep -v banana
	filterReader := &filteringReader{
		source:  inputFile,
		exclude: "banana",
	}

	// Sponge the filtered content back to the same file
	err = RunToFile(filterReader, testFile, 1024)
	inputFile.Close()
	if err != nil {
		t.Fatalf("RunToFile error: %v", err)
	}

	// Verify result
	resultContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	expectedContent := "apple\ncherry\ndate\nelderberry\n"
	if string(resultContent) != expectedContent {
		t.Errorf("Pipeline result incorrect!\nExpected: %q\nGot: %q", expectedContent, string(resultContent))
	}
}

// filteringReader wraps a reader and filters out lines containing the exclude string
type filteringReader struct {
	source  io.Reader
	exclude string
	buffer  []byte
	scanned bool
}

func (f *filteringReader) Read(p []byte) (n int, err error) {
	if !f.scanned {
		// Read all content and filter
		content, err := io.ReadAll(f.source)
		if err != nil {
			return 0, err
		}

		var filtered bytes.Buffer
		for _, line := range strings.Split(string(content), "\n") {
			if line != "" && !strings.Contains(line, f.exclude) {
				filtered.WriteString(line + "\n")
			}
		}
		f.buffer = filtered.Bytes()
		f.scanned = true
	}

	if len(f.buffer) == 0 {
		return 0, io.EOF
	}

	n = copy(p, f.buffer)
	f.buffer = f.buffer[n:]
	return n, nil
}

// TestWithoutSponge_FileTruncated demonstrates why sponge is needed.
// Without sponge, writing to the same file while reading would truncate it.
func TestWithoutSponge_FileTruncated(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "truncate.txt")

	// Create initial file with content
	originalContent := "This content would be lost without sponge!\n"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open file for reading
	inputFile, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer inputFile.Close()

	// Open the SAME file for writing (truncates it!)
	outputFile, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Try to copy - this will fail because file was truncated
	_, _ = io.Copy(outputFile, inputFile)
	outputFile.Close()

	// Read the result
	resultContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	// The file should be empty or have very little content because
	// os.Create truncated it before we could read
	if string(resultContent) == originalContent {
		t.Error("Expected file to be truncated without sponge, but content was preserved")
	}
}

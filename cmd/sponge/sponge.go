package sponge

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	MaxSize string `short:"m" help:"Maximum size to keep in memory before buffering to a temp file (e.g., 10m, 1g)" default:"10m"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "sponge",
		Short:       "Soak up stdin and release to stdout after EOF",
		Long:        "Reads all input from stdin until EOF, storing it in memory. If the input exceeds the max-size, it buffers to a temporary file. Only after all input is read does it output to stdout. Useful for in-place file modifications in pipelines.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			maxBytes, err := common.ParseSize(params.MaxSize)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sponge: invalid max-size: %v\n", err)
				os.Exit(1)
			}
			if err := Run(os.Stdin, os.Stdout, maxBytes); err != nil {
				fmt.Fprintf(os.Stderr, "sponge: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

// Buffer absorbs all data from a reader, buffering in memory up to maxSize,
// then spilling to a temporary file if exceeded. It can then be read from
// or written to a destination.
type Buffer struct {
	maxSize      int64
	memBuffer    []byte
	tempFile     *os.File
	tempWriter   *bufio.Writer
	usingTempFile bool
	totalSize    int64
	readPos      int64
}

// NewBuffer creates a new sponge buffer with the given max memory size.
func NewBuffer(maxSize int64) *Buffer {
	return &Buffer{maxSize: maxSize}
}

// ReadFrom reads all data from r until EOF, buffering it.
// Implements io.ReaderFrom.
func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	bufReader := bufio.NewReader(r)
	chunk := make([]byte, 32*1024) // 32KB chunks

	for {
		n, err := bufReader.Read(chunk)
		if n > 0 {
			if werr := b.write(chunk[:n]); werr != nil {
				return b.totalSize, werr
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return b.totalSize, fmt.Errorf("failed to read input: %w", err)
		}
	}

	// Prepare for reading if using temp file
	if b.usingTempFile {
		if err := b.tempWriter.Flush(); err != nil {
			return b.totalSize, fmt.Errorf("failed to flush temp file: %w", err)
		}
		if _, err := b.tempFile.Seek(0, 0); err != nil {
			return b.totalSize, fmt.Errorf("failed to seek temp file: %w", err)
		}
	}

	return b.totalSize, nil
}

// write appends data to the buffer, switching to temp file if needed.
func (b *Buffer) write(data []byte) error {
	b.totalSize += int64(len(data))

	if !b.usingTempFile {
		if b.totalSize > b.maxSize {
			// Switch to temp file
			var err error
			b.tempFile, err = os.CreateTemp("", "sponge-*")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			b.tempWriter = bufio.NewWriter(b.tempFile)
			b.usingTempFile = true

			// Write existing memory buffer to temp file
			if len(b.memBuffer) > 0 {
				if _, err := b.tempWriter.Write(b.memBuffer); err != nil {
					return fmt.Errorf("failed to write to temp file: %w", err)
				}
				b.memBuffer = nil // Free memory
			}

			// Write current data to temp file
			if _, err := b.tempWriter.Write(data); err != nil {
				return fmt.Errorf("failed to write to temp file: %w", err)
			}
		} else {
			b.memBuffer = append(b.memBuffer, data...)
		}
	} else {
		if _, err := b.tempWriter.Write(data); err != nil {
			return fmt.Errorf("failed to write to temp file: %w", err)
		}
	}

	return nil
}

// WriteTo writes all buffered data to w.
// Implements io.WriterTo.
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	if b.usingTempFile {
		return io.Copy(w, b.tempFile)
	}
	n, err := w.Write(b.memBuffer)
	return int64(n), err
}

// Close cleans up any temporary files.
func (b *Buffer) Close() error {
	if b.tempFile != nil {
		name := b.tempFile.Name()
		b.tempFile.Close()
		return os.Remove(name)
	}
	return nil
}

// Size returns the total size of buffered data.
func (b *Buffer) Size() int64 {
	return b.totalSize
}

// UsingTempFile returns true if the buffer spilled to a temp file.
func (b *Buffer) UsingTempFile() bool {
	return b.usingTempFile
}

// Run reads all input from reader, buffering to memory or a temp file if it exceeds maxSize,
// then writes all content to writer.
func Run(reader io.Reader, writer io.Writer, maxSize int64) error {
	buf := NewBuffer(maxSize)
	defer buf.Close()

	if _, err := buf.ReadFrom(reader); err != nil {
		return err
	}

	if _, err := buf.WriteTo(writer); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// RunToFile reads all input from reader and writes to the specified file path.
// This is the classic sponge use case: cat file | transform | sponge file
func RunToFile(reader io.Reader, filePath string, maxSize int64) error {
	buf := NewBuffer(maxSize)
	defer buf.Close()

	if _, err := buf.ReadFrom(reader); err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if _, err := buf.WriteTo(f); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

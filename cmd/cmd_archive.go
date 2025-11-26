package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/mholt/archives"
	"github.com/spf13/cobra"
)

// ArchiveCreateParams holds parameters for archive creation
type ArchiveCreateParams struct {
	Output  string   `short:"o" help:"Output archive file name (format auto-detected from extension)"`
	Files   []string `pos:"true" optional:"true" help:"Files and directories to archive"`
	Verbose bool     `short:"v" optional:"true" help:"Verbose output - list files as they are added"`
	Format  string   `short:"f" optional:"true" help:"Archive format (tar, tar.gz, tar.bz2, tar.xz, tar.zst, zip, 7z). Overrides extension detection."`
}

// ArchiveExtractParams holds parameters for archive extraction
type ArchiveExtractParams struct {
	Archive  string `pos:"true" help:"Archive file to extract"`
	Output   string `short:"o" optional:"true" help:"Output directory (default: current directory)" default:"."`
	Verbose  bool   `short:"v" optional:"true" help:"Verbose output - list files as they are extracted"`
	Password string `short:"p" optional:"true" help:"Password for encrypted archives (7z, rar)"`
}

// ArchiveListParams holds parameters for listing archive contents
type ArchiveListParams struct {
	Archive  string `pos:"true" help:"Archive file to list"`
	Long     bool   `short:"l" optional:"true" help:"Long listing format (show size and permissions)"`
	Password string `short:"p" optional:"true" help:"Password for encrypted archives (7z, rar)"`
}

func ArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Create and extract archive files",
		Long: `Create, extract, and list archive files in various formats.

Supported formats:
  - tar         Plain tar archive
  - tar.gz, tgz Gzip-compressed tar
  - tar.bz2     Bzip2-compressed tar
  - tar.xz      XZ-compressed tar
  - tar.zst     Zstd-compressed tar
  - tar.lz4     LZ4-compressed tar
  - zip         ZIP archive
  - 7z          7-Zip archive (extract only, password supported)
  - rar         RAR archive (extract only, password supported)

The format is auto-detected from the file extension, or can be specified explicitly.
Password-protected 7z and rar archives can be extracted using the -p flag.`,
	}

	cmd.AddCommand(archiveCreateCmd())
	cmd.AddCommand(archiveExtractCmd())
	cmd.AddCommand(archiveListCmd())

	return cmd
}

func archiveCreateCmd() *cobra.Command {
	return boa.CmdT[ArchiveCreateParams]{
		Use:   "create",
		Short: "Create an archive from files and directories",
		Long: `Create an archive file from the specified files and directories.

The archive format is determined by the output file extension, or can be
specified explicitly with the --format flag.

Examples:
  tofu archive create -o backup.tar.gz file1.txt dir1/
  tofu archive create -o project.zip src/ README.md
  tofu archive create -f tar.zst -o backup.tar.zst data/`,
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *ArchiveCreateParams, cmd *cobra.Command, args []string) {
			if params.Output == "" {
				fmt.Fprintln(os.Stderr, "archive: output file required (-o)")
				os.Exit(1)
			}
			if len(params.Files) == 0 {
				fmt.Fprintln(os.Stderr, "archive: no files specified")
				os.Exit(1)
			}
			if err := runArchiveCreate(params); err != nil {
				fmt.Fprintf(os.Stderr, "archive: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func archiveExtractCmd() *cobra.Command {
	return boa.CmdT[ArchiveExtractParams]{
		Use:   "extract",
		Short: "Extract files from an archive",
		Long: `Extract all files from an archive to the specified directory.

The archive format is auto-detected from the file contents.

Examples:
  tofu archive extract backup.tar.gz
  tofu archive extract -o /tmp/output project.zip
  tofu archive extract -v archive.7z`,
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *ArchiveExtractParams, cmd *cobra.Command, args []string) {
			if params.Archive == "" {
				fmt.Fprintln(os.Stderr, "archive: archive file required")
				os.Exit(1)
			}
			if err := runArchiveExtract(params); err != nil {
				fmt.Fprintf(os.Stderr, "archive: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func archiveListCmd() *cobra.Command {
	return boa.CmdT[ArchiveListParams]{
		Use:   "list",
		Short: "List contents of an archive",
		Long: `List all files contained in an archive.

Examples:
  tofu archive list backup.tar.gz
  tofu archive list -l project.zip`,
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *ArchiveListParams, cmd *cobra.Command, args []string) {
			if params.Archive == "" {
				fmt.Fprintln(os.Stderr, "archive: archive file required")
				os.Exit(1)
			}
			if err := runArchiveList(params); err != nil {
				fmt.Fprintf(os.Stderr, "archive: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runArchiveCreate(params *ArchiveCreateParams) error {
	ctx := context.Background()

	// Determine the archive format
	format, err := getArchiveFormat(params.Output, params.Format)
	if err != nil {
		return err
	}

	archiver, ok := format.(archives.Archiver)
	if !ok {
		return fmt.Errorf("format does not support archive creation")
	}

	// Build the file list
	fileMap := make(map[string]string)
	for _, path := range params.Files {

		// Check if path exists
		_, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("cannot access %s: %w", path, err)
		}

		fileMap[path] = ""
	}

	files, err := archives.FilesFromDisk(ctx, nil, fileMap)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	// Create output file
	outFile, err := os.Create(params.Output)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer outFile.Close()

	// If verbose, wrap with a counting writer
	var output io.Writer = outFile
	if params.Verbose {
		for _, f := range files {
			fmt.Printf("a %s\n", f.NameInArchive)
		}
	}

	// Create the archive
	if err := archiver.Archive(ctx, output, files); err != nil {
		os.Remove(params.Output) // Clean up partial file
		return fmt.Errorf("failed to create archive: %w", err)
	}

	return nil
}

func runArchiveExtract(params *ArchiveExtractParams) error {
	ctx := context.Background()

	// Open the archive file
	archiveFile, err := os.Open(params.Archive)
	if err != nil {
		return fmt.Errorf("cannot open archive: %w", err)
	}
	defer archiveFile.Close()

	// Identify the format
	format, reader, err := archives.Identify(ctx, params.Archive, archiveFile)
	if err != nil {
		return fmt.Errorf("cannot identify archive format: %w", err)
	}

	// Apply password to formats that support it
	if params.Password != "" {
		switch f := format.(type) {
		case archives.SevenZip:
			f.Password = params.Password
			format = f
		case archives.Rar:
			f.Password = params.Password
			format = f
		}
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("format does not support extraction")
	}

	// Create output directory if needed
	if params.Output != "." {
		if err := os.MkdirAll(params.Output, 0755); err != nil {
			return fmt.Errorf("cannot create output directory: %w", err)
		}
	}

	// For formats that need seeking (zip, 7z), we need to use the file directly
	var archiveReader io.Reader = reader
	switch format.(type) {
	case archives.Zip, archives.SevenZip:
		// These formats need the original file for seeking
		archiveFile.Seek(0, io.SeekStart)
		archiveReader = archiveFile
	}

	// Extract files
	err = extractor.Extract(ctx, archiveReader, func(ctx context.Context, f archives.FileInfo) error {
		// Sanitize the path
		destPath := filepath.Join(params.Output, filepath.Clean(f.NameInArchive))

		// Security check: ensure we're not writing outside the output directory
		if !strings.HasPrefix(destPath, filepath.Clean(params.Output)+string(filepath.Separator)) &&
			destPath != filepath.Clean(params.Output) {
			return fmt.Errorf("invalid file path: %s", f.NameInArchive)
		}

		if params.Verbose {
			fmt.Printf("x %s\n", f.NameInArchive)
		}

		// Handle directories
		if f.IsDir() {
			return os.MkdirAll(destPath, f.Mode())
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Handle symlinks
		if f.Mode()&os.ModeSymlink != 0 {
			return os.Symlink(f.LinkTarget, destPath)
		}

		// Extract regular file
		outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(outFile, srcFile)
		return err
	})

	return err
}

func runArchiveList(params *ArchiveListParams) error {
	ctx := context.Background()

	// Open the archive file
	archiveFile, err := os.Open(params.Archive)
	if err != nil {
		return fmt.Errorf("cannot open archive: %w", err)
	}
	defer archiveFile.Close()

	// Identify the format
	format, reader, err := archives.Identify(ctx, params.Archive, archiveFile)
	if err != nil {
		return fmt.Errorf("cannot identify archive format: %w", err)
	}

	// Apply password to formats that support it
	if params.Password != "" {
		switch f := format.(type) {
		case archives.SevenZip:
			f.Password = params.Password
			format = f
		case archives.Rar:
			f.Password = params.Password
			format = f
		}
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("format does not support listing")
	}

	// For formats that need seeking (zip, 7z), we need to use the file directly
	var archiveReader io.Reader = reader
	switch format.(type) {
	case archives.Zip, archives.SevenZip:
		archiveFile.Seek(0, io.SeekStart)
		archiveReader = archiveFile
	}

	// List files
	err = extractor.Extract(ctx, archiveReader, func(ctx context.Context, f archives.FileInfo) error {
		if params.Long {
			mode := f.Mode().String()
			size := f.Size()
			name := f.NameInArchive
			if f.IsDir() {
				name += "/"
			}
			fmt.Printf("%s %10d  %s\n", mode, size, name)
		} else {
			name := f.NameInArchive
			if f.IsDir() {
				name += "/"
			}
			fmt.Println(name)
		}
		return nil
	})

	return err
}

func getArchiveFormat(filename, formatOverride string) (archives.Format, error) {
	// If format is explicitly specified, use it
	if formatOverride != "" {
		return parseFormatString(formatOverride)
	}

	// Otherwise, detect from filename
	return parseFormatFromExtension(filename)
}

func parseFormatString(format string) (archives.Format, error) {
	format = strings.ToLower(format)

	switch format {
	case "tar":
		return archives.Tar{}, nil
	case "tar.gz", "tgz":
		return archives.CompressedArchive{
			Archival:    archives.Tar{},
			Compression: archives.Gz{},
		}, nil
	case "tar.bz2", "tbz2", "tbz":
		return archives.CompressedArchive{
			Archival:    archives.Tar{},
			Compression: archives.Bz2{},
		}, nil
	case "tar.xz", "txz":
		return archives.CompressedArchive{
			Archival:    archives.Tar{},
			Compression: archives.Xz{},
		}, nil
	case "tar.zst", "tar.zstd":
		return archives.CompressedArchive{
			Archival:    archives.Tar{},
			Compression: archives.Zstd{},
		}, nil
	case "tar.lz4":
		return archives.CompressedArchive{
			Archival:    archives.Tar{},
			Compression: archives.Lz4{},
		}, nil
	case "tar.br", "tar.brotli":
		return archives.CompressedArchive{
			Archival:    archives.Tar{},
			Compression: archives.Brotli{},
		}, nil
	case "zip":
		return archives.Zip{}, nil
	case "7z":
		return archives.SevenZip{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func parseFormatFromExtension(filename string) (archives.Format, error) {
	lower := strings.ToLower(filename)

	// Check compound extensions first
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return parseFormatString("tar.gz")
	}
	if strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2") || strings.HasSuffix(lower, ".tbz") {
		return parseFormatString("tar.bz2")
	}
	if strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz") {
		return parseFormatString("tar.xz")
	}
	if strings.HasSuffix(lower, ".tar.zst") || strings.HasSuffix(lower, ".tar.zstd") {
		return parseFormatString("tar.zst")
	}
	if strings.HasSuffix(lower, ".tar.lz4") {
		return parseFormatString("tar.lz4")
	}
	if strings.HasSuffix(lower, ".tar.br") {
		return parseFormatString("tar.br")
	}

	// Simple extensions
	ext := strings.TrimPrefix(filepath.Ext(lower), ".")
	switch ext {
	case "tar":
		return parseFormatString("tar")
	case "zip":
		return parseFormatString("zip")
	case "7z":
		return parseFormatString("7z")
	default:
		return nil, fmt.Errorf("cannot determine format from extension: %s", ext)
	}
}

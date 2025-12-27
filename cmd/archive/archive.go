package archive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/mholt/archives"
	"github.com/spf13/cobra"
	"github.com/yeka/zip"
)

// CreateParams holds parameters for archive creation
type CreateParams struct {
	Output     string   `short:"o" help:"Output archive file name (format auto-detected from extension)"`
	Files      []string `pos:"true" optional:"true" help:"Files and directories to archive"`
	Verbose    bool     `short:"v" optional:"true" help:"Verbose output - list files as they are added"`
	Format     string   `short:"f" optional:"true" help:"Archive format (tar, tar.gz, tar.bz2, tar.xz, tar.zst, zip, 7z). Overrides extension detection."`
	Password   string   `short:"p" optional:"true" help:"Password for encrypted ZIP archives"`
	Encryption string   `short:"e" optional:"true" help:"Encryption method for ZIP: legacy (insecure), aes128, aes192, aes256 (default: aes256)" default:"aes256"`
}

// ExtractParams holds parameters for archive extraction
type ExtractParams struct {
	Archive  string `pos:"true" help:"Archive file to extract"`
	Output   string `short:"o" optional:"true" help:"Output directory (default: current directory)" default:"."`
	Verbose  bool   `short:"v" optional:"true" help:"Verbose output - list files as they are extracted"`
	Password string `short:"p" optional:"true" help:"Password for encrypted archives (zip, 7z, rar)"`
}

// ListParams holds parameters for listing archive contents
type ListParams struct {
	Archive  string `pos:"true" help:"Archive file to list"`
	Long     bool   `short:"l" optional:"true" help:"Long listing format (show size and permissions)"`
	Password string `short:"p" optional:"true" help:"Password for encrypted archives (zip, 7z, rar)"`
}

func Cmd() *cobra.Command {
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
  - zip         ZIP archive (password supported with AES encryption)
  - 7z          7-Zip archive (extract only, password supported)
  - rar         RAR archive (extract only, password supported)

The format is auto-detected from the file extension, or can be specified explicitly.
Password-protected zip, 7z, and rar archives can be extracted using the -p flag.
ZIP archives can be created with password protection using the -p flag (AES encryption).`,
	}

	cmd.AddCommand(createCmd())
	cmd.AddCommand(extractCmd())
	cmd.AddCommand(listCmd())

	return cmd
}

func createCmd() *cobra.Command {
	return boa.CmdT[CreateParams]{
		Use:   "create",
		Short: "Create an archive from files and directories",
		Long: `Create an archive file from the specified files and directories.

The archive format is determined by the output file extension, or can be
specified explicitly with the --format flag.

ZIP archives can be encrypted using the -p (password) and -e (encryption) flags.
Supported encryption methods: legacy (insecure, for compatibility), aes128, aes192, aes256 (default).

Examples:
  tofu archive create -o backup.tar.gz file1.txt dir1/
  tofu archive create -o project.zip src/ README.md
  tofu archive create -f tar.zst -o backup.tar.zst data/
  tofu archive create -o secret.zip -p mypassword file.txt
  tofu archive create -o secret.zip -p mypassword -e aes128 file.txt
  tofu archive create -o compat.zip -p mypassword -e legacy file.txt`,
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *CreateParams, cmd *cobra.Command) error {
			cmd.Aliases = []string{"c"}
			return nil
		},
		RunFunc: func(params *CreateParams, cmd *cobra.Command, args []string) {
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

func extractCmd() *cobra.Command {
	return boa.CmdT[ExtractParams]{
		Use:   "extract",
		Short: "Extract files from an archive",
		Long: `Extract all files from an archive to the specified directory.

The archive format is auto-detected from the file contents.
For encrypted archives (zip, 7z, rar), use the -p flag to specify the password.

Examples:
  tofu archive extract backup.tar.gz
  tofu archive extract -o /tmp/output project.zip
  tofu archive extract -v archive.7z
  tofu archive extract -p mypassword secret.zip`,
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *ExtractParams, cmd *cobra.Command) error {
			cmd.Aliases = []string{"x"}
			return nil
		},
		RunFunc: func(params *ExtractParams, cmd *cobra.Command, args []string) {
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

func listCmd() *cobra.Command {
	return boa.CmdT[ListParams]{
		Use:   "list",
		Short: "List contents of an archive",
		Long: `List all files contained in an archive.

For encrypted archives (zip, 7z, rar), use the -p flag to specify the password.

Examples:
  tofu archive list backup.tar.gz
  tofu archive list -l project.zip
  tofu archive list -p mypassword secret.zip`,
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *ListParams, cmd *cobra.Command) error {
			cmd.Aliases = []string{"l", "ls"}
			return nil
		},
		RunFunc: func(params *ListParams, cmd *cobra.Command, args []string) {
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

func runArchiveCreate(params *CreateParams) error {
	ctx := context.Background()

	// Determine the archive format
	format, err := getArchiveFormat(params.Output, params.Format)
	if err != nil {
		return err
	}

	// Use encrypted ZIP writer when password provided for zip format
	if params.Password != "" {
		if _, isZip := format.(archives.Zip); isZip {
			return createEncryptedZip(params)
		}
		return fmt.Errorf("password encryption is only supported for ZIP format")
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

func runArchiveExtract(params *ExtractParams) error {
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
		case archives.Zip:
			// Use yeka/zip for encrypted ZIP extraction
			archiveFile.Close()
			return extractEncryptedZip(params)
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
	absOutputRootDir, err := filepath.Abs(params.Output)
	if err != nil {
		return fmt.Errorf("invalid output directory: %s", params.Output)
	}
	err = extractor.Extract(ctx, archiveReader, func(ctx context.Context, f archives.FileInfo) error {
		// Sanitize the path
		destPath := filepath.Join(absOutputRootDir, filepath.Clean(f.NameInArchive))
		destPathAbs, err := filepath.Abs(destPath)
		if err != nil {
			return fmt.Errorf("invalid file path: %s", f.NameInArchive)
		}

		// Security check: ensure we're not writing outside the output directory
		if !strings.HasPrefix(destPathAbs, filepath.Clean(absOutputRootDir)) &&
			destPathAbs != filepath.Clean(absOutputRootDir) {
			return fmt.Errorf("invalid file path: %s", f.NameInArchive)
		}

		if params.Verbose {
			fmt.Printf("x %s\n", f.NameInArchive)
		}

		// Handle directories
		if f.IsDir() {
			return os.MkdirAll(destPathAbs, f.Mode())
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPathAbs), 0755); err != nil {
			return err
		}

		// Handle symlinks
		if f.Mode()&os.ModeSymlink != 0 {
			return os.Symlink(f.LinkTarget, destPath)
		}

		// Extract regular file
		outFile, err := os.OpenFile(destPathAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
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

func runArchiveList(params *ListParams) error {
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
		case archives.Zip:
			// Use yeka/zip for encrypted ZIP listing
			archiveFile.Close()
			return listEncryptedZip(params)
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

func parseEncryptionMethod(method string) (zip.EncryptionMethod, error) {
	switch strings.ToLower(method) {
	case "legacy", "zipcrypto":
		return zip.StandardEncryption, nil
	case "aes128":
		return zip.AES128Encryption, nil
	case "aes192":
		return zip.AES192Encryption, nil
	case "aes256", "":
		return zip.AES256Encryption, nil
	default:
		return 0, fmt.Errorf("invalid encryption method: %s (use legacy, aes128, aes192, or aes256)", method)
	}
}

func createEncryptedZip(params *CreateParams) error {
	encMethod, err := parseEncryptionMethod(params.Encryption)
	if err != nil {
		return err
	}

	// Create output file
	outFile, err := os.Create(params.Output)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	// Process each input file/directory
	for _, inputPath := range params.Files {
		info, err := os.Lstat(inputPath)
		if err != nil {
			os.Remove(params.Output)
			return fmt.Errorf("cannot access %s: %w", inputPath, err)
		}

		if info.IsDir() {
			// Walk directory - compute relative paths from the directory's parent
			baseDir := filepath.Dir(inputPath)
			err = filepath.Walk(inputPath, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				// Compute name relative to the base directory
				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					relPath = filepath.Base(path)
				}
				return addFileToEncryptedZip(zw, path, relPath, fi, params.Password, encMethod, params.Verbose)
			})
			if err != nil {
				os.Remove(params.Output)
				return fmt.Errorf("failed to add directory %s: %w", inputPath, err)
			}
		} else {
			// Single file - use just the base name
			nameInArchive := filepath.Base(inputPath)
			if err := addFileToEncryptedZip(zw, inputPath, nameInArchive, info, params.Password, encMethod, params.Verbose); err != nil {
				os.Remove(params.Output)
				return fmt.Errorf("failed to add file %s: %w", inputPath, err)
			}
		}
	}

	return nil
}

func addFileToEncryptedZip(zw *zip.Writer, path string, nameInArchive string, info os.FileInfo, password string, encMethod zip.EncryptionMethod, verbose bool) error {
	if verbose {
		fmt.Printf("a %s\n", nameInArchive)
	}

	// Handle directories
	if info.IsDir() {
		_, err := zw.Create(nameInArchive + "/")
		return err
	}

	// Handle symlinks - store them as regular entries with link target as content
	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(path)
		if err != nil {
			return err
		}
		w, err := zw.Encrypt(nameInArchive, password, encMethod)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(linkTarget))
		return err
	}

	// Regular file - encrypt it
	w, err := zw.Encrypt(nameInArchive, password, encMethod)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

func extractEncryptedZip(params *ExtractParams) error {
	zr, err := zip.OpenReader(params.Archive)
	if err != nil {
		return fmt.Errorf("cannot open archive: %w", err)
	}
	defer zr.Close()

	// Create output directory if needed
	absOutputRootDir, err := filepath.Abs(params.Output)
	if err != nil {
		return fmt.Errorf("invalid output directory: %s", params.Output)
	}
	if params.Output != "." {
		if err := os.MkdirAll(absOutputRootDir, 0755); err != nil {
			return fmt.Errorf("cannot create output directory: %w", err)
		}
	}

	for _, f := range zr.File {
		// Set password if file is encrypted
		if f.IsEncrypted() {
			f.SetPassword(params.Password)
		}

		// Sanitize the path
		destPath := filepath.Join(absOutputRootDir, filepath.Clean(f.Name))
		destPathAbs, err := filepath.Abs(destPath)
		if err != nil {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		// Security check: ensure we're not writing outside the output directory
		if !strings.HasPrefix(destPathAbs, filepath.Clean(absOutputRootDir)) &&
			destPathAbs != filepath.Clean(absOutputRootDir) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if params.Verbose {
			fmt.Printf("x %s\n", f.Name)
		}

		// Handle directories
		if f.FileInfo().IsDir() {
			// Use 0755 for directories to ensure they're writable
			mode := f.Mode()
			if mode == 0 {
				mode = 0755
			}
			if err := os.MkdirAll(destPathAbs, mode|0700); err != nil {
				return err
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPathAbs), 0755); err != nil {
			return err
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open file in archive: %s: %w", f.Name, err)
		}

		outFile, err := os.OpenFile(destPathAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func listEncryptedZip(params *ListParams) error {
	zr, err := zip.OpenReader(params.Archive)
	if err != nil {
		return fmt.Errorf("cannot open archive: %w", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		// Set password if file is encrypted (needed to read file info for some archives)
		if f.IsEncrypted() && params.Password != "" {
			f.SetPassword(params.Password)
		}

		if params.Long {
			mode := f.Mode().String()
			size := f.UncompressedSize64
			name := f.Name
			if f.FileInfo().IsDir() {
				name += "/"
			}
			fmt.Printf("%s %10d  %s\n", mode, size, name)
		} else {
			name := f.Name
			if f.FileInfo().IsDir() {
				name += "/"
			}
			fmt.Println(name)
		}
	}

	return nil
}

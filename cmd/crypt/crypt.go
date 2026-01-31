package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

const (
	// OpenSSL format constants
	opensslSaltHeader = "Salted__"
	opensslSaltSize   = 8
	opensslKeySize    = 32     // AES-256
	opensslIVSize     = 16     // AES block size
	opensslIterations = 600000 // PBKDF2 iterations (modern recommendation)
)

type EncryptParams struct {
	Files    []string `pos:"true" help:"Files to encrypt"`
	Output   string   `short:"o" optional:"true" help:"Output file (only valid with single input file)"`
	Password string   `short:"p" optional:"true" help:"Encryption password (will prompt if not provided)"`
	Format   string   `short:"f" optional:"true" help:"Output format: age (default, modern), openssl (compatible with openssl enc)." default:"age" alts:"age,openssl"`
	Keep     bool     `short:"k" optional:"true" help:"Keep original files after encryption." default:"false"`
	Force    bool     `short:"F" optional:"true" help:"Overwrite output files if they exist." default:"false"`
	Verbose  bool     `short:"v" optional:"true" help:"Verbose output."`
}

type DecryptParams struct {
	Files    []string `pos:"true" help:"Files to decrypt"`
	Output   string   `short:"o" optional:"true" help:"Output file (only valid with single input file)"`
	Password string   `short:"p" optional:"true" help:"Decryption password (will prompt if not provided)"`
	Format   string   `short:"f" optional:"true" help:"Input format: auto (default), age, openssl." default:"auto" alts:"auto,age,openssl"`
	Keep     bool     `short:"k" optional:"true" help:"Keep encrypted files after decryption." default:"false"`
	Force    bool     `short:"F" optional:"true" help:"Overwrite output files if they exist." default:"false"`
	Verbose  bool     `short:"v" optional:"true" help:"Verbose output."`
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crypt",
		Short: "Encrypt and decrypt files",
		Long: `Encrypt and decrypt files using modern authenticated encryption.

Supported formats:
  - age       Modern encryption (default). Compatible with the 'age' CLI tool.
              Uses scrypt for key derivation, ChaCha20-Poly1305 for encryption.
  - openssl   Compatible with 'openssl enc -aes-256-cbc -pbkdf2'.
              Uses PBKDF2 for key derivation, AES-256-CBC for encryption.

The 'age' format is recommended for security. Use 'openssl' for compatibility
with systems that only have OpenSSL available.

Examples:
  tofu crypt encrypt secret.txt                    # age format (default)
  tofu crypt encrypt -f openssl secret.txt         # openssl compatible
  tofu crypt decrypt secret.txt.age                # auto-detects format
  tofu crypt decrypt -p mypassword secret.txt.age

Interoperability:
  # Encrypt with tofu, decrypt with age
  tofu crypt encrypt -p secret file.txt
  age -d -o file.txt file.txt.age

  # Encrypt with age, decrypt with tofu
  age -p -o file.age file.txt
  tofu crypt decrypt file.age

  # Encrypt with tofu, decrypt with openssl
  tofu crypt encrypt -f openssl file.txt
  openssl enc -d -aes-256-cbc -pbkdf2 -in file.txt.enc -out file.txt

  # Encrypt with openssl, decrypt with tofu
  openssl enc -aes-256-cbc -pbkdf2 -in file.txt -out file.enc
  tofu crypt decrypt -f openssl file.enc`,
	}

	cmd.AddCommand(encryptCmd())
	cmd.AddCommand(decryptCmd())

	return cmd
}

func encryptCmd() *cobra.Command {
	return boa.CmdT[EncryptParams]{
		Use:   "encrypt",
		Short: "Encrypt files",
		Long: `Encrypt one or more files.

The password can be provided via -p flag or will be prompted interactively.
Default output extension is .age (or .enc for openssl format).

Examples:
  tofu crypt encrypt secret.txt
  tofu crypt encrypt -p mypassword document.pdf
  tofu crypt encrypt -f openssl -o backup.enc important.txt
  tofu crypt encrypt -k file1.txt file2.txt`,
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *EncryptParams, cmd *cobra.Command) error {
			cmd.Aliases = []string{"e", "enc"}
			return nil
		},
		RunFunc: func(params *EncryptParams, cmd *cobra.Command, args []string) {
			if err := runEncrypt(params); err != nil {
				fmt.Fprintf(os.Stderr, "crypt: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func decryptCmd() *cobra.Command {
	return boa.CmdT[DecryptParams]{
		Use:   "decrypt",
		Short: "Decrypt files",
		Long: `Decrypt one or more encrypted files.

The password can be provided via -p flag or will be prompted interactively.
Format is auto-detected by default (age files start with "age-encryption.org",
openssl files start with "Salted__").

Examples:
  tofu crypt decrypt secret.txt.age
  tofu crypt decrypt -p mypassword document.pdf.enc
  tofu crypt decrypt -f openssl legacy.enc
  tofu crypt decrypt -k *.age`,
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *DecryptParams, cmd *cobra.Command) error {
			cmd.Aliases = []string{"d", "dec"}
			return nil
		},
		RunFunc: func(params *DecryptParams, cmd *cobra.Command, args []string) {
			if err := runDecrypt(params); err != nil {
				fmt.Fprintf(os.Stderr, "crypt: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runEncrypt(params *EncryptParams) error {
	if len(params.Files) == 0 {
		return errors.New("no files specified")
	}

	if params.Output != "" && len(params.Files) > 1 {
		return errors.New("-o can only be used with a single input file")
	}

	// Validate format
	format := strings.ToLower(params.Format)
	if format != "age" && format != "openssl" {
		return fmt.Errorf("unknown format: %s (use age or openssl)", params.Format)
	}

	// Get password
	password, err := getPassword(params.Password, true)
	if err != nil {
		return err
	}

	// Determine file extension
	ext := ".age"
	if format == "openssl" {
		ext = ".enc"
	}

	for _, inputPath := range params.Files {
		outputPath := params.Output
		if outputPath == "" {
			outputPath = inputPath + ext
		}

		// Check if output exists
		if !params.Force {
			if _, err := os.Stat(outputPath); err == nil {
				return fmt.Errorf("output file already exists: %s (use -F to overwrite)", outputPath)
			}
		}

		if params.Verbose {
			fmt.Printf("encrypting %s -> %s (%s format)\n", inputPath, outputPath, format)
		}

		var encryptErr error
		if format == "age" {
			encryptErr = encryptFileAge(inputPath, outputPath, password)
		} else {
			encryptErr = encryptFileOpenSSL(inputPath, outputPath, password)
		}

		if encryptErr != nil {
			return fmt.Errorf("failed to encrypt %s: %w", inputPath, encryptErr)
		}

		// Remove original if not keeping
		if !params.Keep {
			if err := os.Remove(inputPath); err != nil {
				return fmt.Errorf("failed to remove original file %s: %w", inputPath, err)
			}
		}
	}

	return nil
}

func runDecrypt(params *DecryptParams) error {
	if len(params.Files) == 0 {
		return errors.New("no files specified")
	}

	if params.Output != "" && len(params.Files) > 1 {
		return errors.New("-o can only be used with a single input file")
	}

	// Get password
	password, err := getPassword(params.Password, false)
	if err != nil {
		return err
	}

	for _, inputPath := range params.Files {
		// Detect or use specified format
		format := strings.ToLower(params.Format)
		if format == "auto" {
			detected, err := detectFormat(inputPath)
			if err != nil {
				return fmt.Errorf("failed to detect format for %s: %w", inputPath, err)
			}
			format = detected
		}

		outputPath := params.Output
		if outputPath == "" {
			outputPath = determineDecryptOutputPath(inputPath, format)
		}

		// Check if output exists
		if !params.Force {
			if _, err := os.Stat(outputPath); err == nil {
				return fmt.Errorf("output file already exists: %s (use -F to overwrite)", outputPath)
			}
		}

		if params.Verbose {
			fmt.Printf("decrypting %s -> %s (%s format)\n", inputPath, outputPath, format)
		}

		var decryptErr error
		if format == "age" {
			decryptErr = decryptFileAge(inputPath, outputPath, password)
		} else if format == "openssl" {
			decryptErr = decryptFileOpenSSL(inputPath, outputPath, password)
		} else {
			return fmt.Errorf("unknown format: %s", format)
		}

		if decryptErr != nil {
			return fmt.Errorf("failed to decrypt %s: %w", inputPath, decryptErr)
		}

		// Remove encrypted file if not keeping
		if !params.Keep {
			if err := os.Remove(inputPath); err != nil {
				return fmt.Errorf("failed to remove encrypted file %s: %w", inputPath, err)
			}
		}
	}

	return nil
}

func determineDecryptOutputPath(inputPath, format string) string {
	// Try to remove known extensions
	for _, ext := range []string{".age", ".enc"} {
		if trimmed, ok := strings.CutSuffix(inputPath, ext); ok {
			return trimmed
		}
	}
	// Fallback: add .dec
	return inputPath + ".dec"
}

func detectFormat(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Read enough bytes to detect format
	header := make([]byte, 32)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}
	header = header[:n]

	// Check for age format (starts with "age-encryption.org/v1")
	if strings.HasPrefix(string(header), "age-encryption.org/") {
		return "age", nil
	}

	// Check for OpenSSL format (starts with "Salted__")
	if strings.HasPrefix(string(header), opensslSaltHeader) {
		return "openssl", nil
	}

	// Default to age if can't detect
	return "age", nil
}

func getPassword(provided string, confirm bool) (string, error) {
	if provided != "" {
		return provided, nil
	}

	// Read from terminal
	fmt.Fprint(os.Stderr, "Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	if len(password) == 0 {
		return "", errors.New("password cannot be empty")
	}

	if confirm {
		fmt.Fprint(os.Stderr, "Confirm password: ")
		confirmPw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("failed to read password confirmation: %w", err)
		}

		if string(password) != string(confirmPw) {
			return "", errors.New("passwords do not match")
		}
	}

	return string(password), nil
}

// ============================================================================
// Age format implementation
// ============================================================================

func encryptFileAge(inputPath, outputPath, password string) error {
	// Read input file
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read input file: %w", err)
	}

	// Create scrypt recipient (for passphrase encryption)
	recipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return fmt.Errorf("failed to create recipient: %w", err)
	}

	// Get original file permissions
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("cannot stat input file: %w", err)
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create output directory: %w", err)
		}
	}

	// Create output file
	outFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer outFile.Close()

	// Create encrypted writer
	w, err := age.Encrypt(outFile, recipient)
	if err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	// Write plaintext
	if _, err := w.Write(plaintext); err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("failed to write encrypted data: %w", err)
	}

	// Close to finalize encryption
	if err := w.Close(); err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("failed to finalize encryption: %w", err)
	}

	return nil
}

func decryptFileAge(inputPath, outputPath, password string) error {
	// Open input file
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("cannot open input file: %w", err)
	}
	defer inFile.Close()

	// Create scrypt identity (for passphrase decryption)
	identity, err := age.NewScryptIdentity(password)
	if err != nil {
		return fmt.Errorf("failed to create identity: %w", err)
	}

	// Create decrypted reader
	r, err := age.Decrypt(inFile, identity)
	if err != nil {
		return errors.New("decryption failed: wrong password or corrupted file")
	}

	// Read all decrypted data
	plaintext, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read decrypted data: %w", err)
	}

	// Get original file permissions
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("cannot stat input file: %w", err)
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create output directory: %w", err)
		}
	}

	// Write output file
	if err := os.WriteFile(outputPath, plaintext, info.Mode()); err != nil {
		return fmt.Errorf("cannot write output file: %w", err)
	}

	return nil
}

// ============================================================================
// OpenSSL format implementation
// Compatible with: openssl enc -aes-256-cbc -pbkdf2 -iter 600000
// ============================================================================

func encryptFileOpenSSL(inputPath, outputPath, password string) error {
	// Read input file
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read input file: %w", err)
	}

	// Generate random salt
	salt := make([]byte, opensslSaltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key and IV using PBKDF2
	key, iv := deriveKeyAndIV([]byte(password), salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Pad plaintext to block size (PKCS7)
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	// Encrypt using CBC mode
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	// Build output: "Salted__" + salt + ciphertext
	output := make([]byte, 0, len(opensslSaltHeader)+opensslSaltSize+len(ciphertext))
	output = append(output, []byte(opensslSaltHeader)...)
	output = append(output, salt...)
	output = append(output, ciphertext...)

	// Get original file permissions
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("cannot stat input file: %w", err)
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create output directory: %w", err)
		}
	}

	// Write output file
	if err := os.WriteFile(outputPath, output, info.Mode()); err != nil {
		return fmt.Errorf("cannot write output file: %w", err)
	}

	return nil
}

func decryptFileOpenSSL(inputPath, outputPath, password string) error {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read input file: %w", err)
	}

	// Verify header
	headerLen := len(opensslSaltHeader) + opensslSaltSize
	if len(data) < headerLen {
		return errors.New("invalid openssl encrypted file: too short")
	}

	if string(data[:len(opensslSaltHeader)]) != opensslSaltHeader {
		return errors.New("invalid openssl encrypted file: missing salt header")
	}

	// Extract salt and ciphertext
	salt := data[len(opensslSaltHeader):headerLen]
	ciphertext := data[headerLen:]

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return errors.New("invalid openssl encrypted file: invalid ciphertext length")
	}

	// Derive key and IV using PBKDF2
	key, iv := deriveKeyAndIV([]byte(password), salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt using CBC mode
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return errors.New("decryption failed: wrong password or corrupted file")
	}

	// Get original file permissions
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("cannot stat input file: %w", err)
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create output directory: %w", err)
		}
	}

	// Write output file
	if err := os.WriteFile(outputPath, plaintext, info.Mode()); err != nil {
		return fmt.Errorf("cannot write output file: %w", err)
	}

	return nil
}

// deriveKeyAndIV derives a key and IV from password and salt using PBKDF2
func deriveKeyAndIV(password, salt []byte) (key, iv []byte) {
	// Derive key+IV material using PBKDF2 with SHA-256
	derived := pbkdf2.Key(password, salt, opensslIterations, opensslKeySize+opensslIVSize, sha256.New)
	return derived[:opensslKeySize], derived[opensslKeySize:]
}

// pkcs7Pad pads data to a multiple of blockSize using PKCS7
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padBytes := make([]byte, padding)
	for i := range padBytes {
		padBytes[i] = byte(padding)
	}
	return append(data, padBytes...)
}

// pkcs7Unpad removes PKCS7 padding
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(data) {
		return nil, errors.New("invalid padding")
	}

	// Verify all padding bytes
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

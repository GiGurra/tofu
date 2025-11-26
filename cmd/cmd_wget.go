package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type WgetParams struct {
	URL        string `pos:"true" help:"URL to download"`
	Output     string `short:"O" optional:"true" help:"Write output to file (use '-' for stdout)"`
	Continue   bool   `short:"c" optional:"true" help:"Resume a partially downloaded file"`
	Quiet      bool   `short:"q" optional:"true" help:"Quiet mode - no progress output"`
	NoProgress bool   `optional:"true" help:"Disable progress bar but show other output"`
	Insecure   bool   `short:"k" optional:"true" help:"Allow insecure server connections when using SSL"`
	Timeout    int    `short:"T" optional:"true" help:"Set timeout in seconds" default:"30"`
	Retries    int    `short:"t" optional:"true" help:"Set number of retries (0 for infinite)" default:"3"`
	UserAgent  string `short:"U" optional:"true" help:"Set User-Agent header"`
	Headers    []string `short:"H" optional:"true" help:"Add custom header(s)"`
}

func WgetCmd() *cobra.Command {
	return boa.CmdT[WgetParams]{
		Use:   "wget",
		Short: "Download files from the web",
		Long: `Download files from URLs with progress indication and resume support.

Features:
  - Auto-detect output filename from URL
  - Progress bar showing download status
  - Resume partially downloaded files with -c
  - Follow redirects automatically

Examples:
  tofu wget https://example.com/file.zip
  tofu wget -O output.zip https://example.com/file.zip
  tofu wget -c https://example.com/large-file.iso
  tofu wget -q https://example.com/file.txt`,
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *WgetParams, cmd *cobra.Command, args []string) {
			if params.URL == "" {
				fmt.Fprintln(os.Stderr, "wget: missing URL")
				os.Exit(1)
			}

			// Auto-detect URL scheme if missing
			if !strings.HasPrefix(params.URL, "http://") && !strings.HasPrefix(params.URL, "https://") {
				params.URL = "https://" + params.URL
			}

			if err := runWget(params); err != nil {
				fmt.Fprintf(os.Stderr, "wget: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runWget(params *WgetParams) error {
	// Determine output filename
	outputFile := params.Output
	if outputFile == "" {
		outputFile = filenameFromURL(params.URL)
		if outputFile == "" {
			outputFile = "index.html"
		}
	}

	writeToStdout := outputFile == "-"

	// Check for existing file (for resume)
	var existingSize int64
	if params.Continue && !writeToStdout {
		if info, err := os.Stat(outputFile); err == nil {
			existingSize = info.Size()
		}
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: time.Duration(params.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	if params.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Retry loop
	var lastErr error
	maxRetries := params.Retries
	if maxRetries == 0 {
		maxRetries = 1000000 // "infinite"
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			if !params.Quiet {
				fmt.Fprintf(os.Stderr, "Retrying (%d/%d)...\n", attempt, params.Retries)
			}
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
		}

		err := downloadFile(client, params, outputFile, existingSize, writeToStdout)
		if err == nil {
			return nil
		}
		lastErr = err

		// Don't retry on certain errors
		if strings.Contains(err.Error(), "404") ||
			strings.Contains(err.Error(), "403") ||
			strings.Contains(err.Error(), "401") {
			return err
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", params.Retries, lastErr)
}

func downloadFile(client *http.Client, params *WgetParams, outputFile string, existingSize int64, writeToStdout bool) error {
	req, err := http.NewRequest("GET", params.URL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Set User-Agent
	if params.UserAgent != "" {
		req.Header.Set("User-Agent", params.UserAgent)
	} else {
		req.Header.Set("User-Agent", "tofu/wget")
	}

	// Add custom headers
	for _, h := range params.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Resume support
	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	if !params.Quiet {
		fmt.Fprintf(os.Stderr, "Downloading: %s\n", params.URL)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	// Handle resume response
	var totalSize int64
	resuming := false
	if resp.StatusCode == http.StatusPartialContent {
		resuming = true
		totalSize = existingSize + resp.ContentLength
		if !params.Quiet {
			fmt.Fprintf(os.Stderr, "Resuming from byte %d\n", existingSize)
		}
	} else if resp.StatusCode == http.StatusOK {
		totalSize = resp.ContentLength
		existingSize = 0 // Server doesn't support resume, start fresh
	} else if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		// File already complete
		if !params.Quiet {
			fmt.Fprintf(os.Stderr, "File already complete\n")
		}
		return nil
	}

	// Open output file
	var out io.Writer
	if writeToStdout {
		out = os.Stdout
	} else {
		flags := os.O_CREATE | os.O_WRONLY
		if resuming {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}
		f, err := os.OpenFile(outputFile, flags, 0644)
		if err != nil {
			return fmt.Errorf("cannot create file: %w", err)
		}
		defer f.Close()
		out = f

		if !params.Quiet {
			fmt.Fprintf(os.Stderr, "Saving to: %s\n", outputFile)
		}
	}

	// Wrap with progress if not quiet and not stdout
	var reader io.Reader = resp.Body
	if !params.Quiet && !params.NoProgress && !writeToStdout {
		reader = &progressReader{
			reader:     resp.Body,
			total:      totalSize,
			downloaded: existingSize,
			lastPrint:  time.Now(),
		}
	}

	// Copy data
	written, err := io.Copy(out, reader)
	if err != nil {
		return fmt.Errorf("download error: %w", err)
	}

	// Print final newline after progress bar
	if !params.Quiet && !params.NoProgress && !writeToStdout {
		fmt.Fprintln(os.Stderr)
	}

	if !params.Quiet {
		fmt.Fprintf(os.Stderr, "Downloaded: %s (%s)\n", outputFile, formatBytes(written+existingSize))
	}

	return nil
}

type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	lastPrint  time.Time
	startTime  time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	if pr.startTime.IsZero() {
		pr.startTime = time.Now()
	}

	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	// Update progress every 100ms
	if time.Since(pr.lastPrint) > 100*time.Millisecond || err == io.EOF {
		pr.printProgress()
		pr.lastPrint = time.Now()
	}

	return n, err
}

func (pr *progressReader) printProgress() {
	var percent float64
	var bar string

	if pr.total > 0 {
		percent = float64(pr.downloaded) / float64(pr.total) * 100
		barWidth := 30
		filled := int(percent / 100 * float64(barWidth))
		bar = strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
	} else {
		bar = strings.Repeat("?", 30)
	}

	// Calculate speed
	elapsed := time.Since(pr.startTime).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(pr.downloaded) / elapsed
	}

	if pr.total > 0 {
		fmt.Fprintf(os.Stderr, "\r[%s] %5.1f%% %s/s %s/%s",
			bar, percent, formatBytes(int64(speed)),
			formatBytes(pr.downloaded), formatBytes(pr.total))
	} else {
		fmt.Fprintf(os.Stderr, "\r[%s] %s %s/s",
			bar, formatBytes(pr.downloaded), formatBytes(int64(speed)))
	}
}

func filenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Get the last path component
	filename := path.Base(u.Path)
	if filename == "" || filename == "/" || filename == "." {
		return ""
	}

	return filename
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseContentRange parses Content-Range header like "bytes 0-499/1234"
func parseContentRange(header string) (start, end, total int64, err error) {
	if header == "" {
		return 0, 0, 0, fmt.Errorf("empty header")
	}

	// Remove "bytes " prefix
	header = strings.TrimPrefix(header, "bytes ")

	// Split range and total
	parts := strings.Split(header, "/")
	if len(parts) != 2 {
		return 0, 0, 0, fmt.Errorf("invalid format")
	}

	// Parse total
	if parts[1] != "*" {
		total, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, 0, err
		}
	}

	// Parse range
	rangeParts := strings.Split(parts[0], "-")
	if len(rangeParts) != 2 {
		return 0, 0, 0, fmt.Errorf("invalid range format")
	}

	start, err = strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	end, err = strconv.ParseInt(rangeParts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return start, end, total, nil
}

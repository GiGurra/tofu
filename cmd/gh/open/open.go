package open

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/GiGurra/cmder"
	"github.com/spf13/cobra"
)

type Params struct {
	Target string `pos:"true" optional:"true" help:"Directory path or URL to open (defaults to current directory)" default:"."`
	Remote string `short:"r" help:"Git remote name to use" default:"origin"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:   "open [dir|url]",
		Short: "Open a GitHub repository in the browser",
		Long: `Open a GitHub repository in the default web browser.

If no argument is provided, opens the current directory's remote.
If a directory is provided, opens that directory's remote.
If a URL is provided, opens that URL directly.`,
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := run(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func run(params *Params) error {
	target := params.Target

	// Check if it's a full URL (with scheme)
	if isURL(target) {
		return openBrowser(target)
	}

	// Check if it's a local path that exists
	absPath, err := filepath.Abs(target)
	if err == nil {
		if info, statErr := os.Stat(absPath); statErr == nil {
			dir := absPath
			if !info.IsDir() {
				dir = filepath.Dir(absPath)
			}
			url, err := getRemoteURL(dir, params.Remote)
			if err != nil {
				return err
			}
			return openBrowser(url)
		}
	}

	// Path doesn't exist - check if it looks like a repo URL without scheme
	if looksLikeRepoURL(target) {
		return openBrowser("https://" + target)
	}

	return fmt.Errorf("path not found and not a valid URL: %s", target)
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func looksLikeRepoURL(s string) bool {
	// Check for repo URLs without scheme (e.g., github.com/owner/repo)
	// Called only after verifying the path doesn't exist locally
	parts := strings.SplitN(s, "/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return false
	}
	domain := parts[0]
	// Domain must look like a hostname (contains a dot, doesn't start with dot)
	return strings.Contains(domain, ".") && !strings.HasPrefix(domain, ".")
}

func getRemoteURL(dir, remote string) (string, error) {
	result := cmder.New("git", "-C", dir, "remote", "get-url", remote).
		WithAttemptTimeout(5 * time.Second).
		Run(context.Background())

	if result.Err != nil {
		return "", fmt.Errorf("failed to get git remote URL: %w", result.Err)
	}

	remoteURL := strings.TrimSpace(result.StdOut)
	if remoteURL == "" {
		return "", fmt.Errorf("no remote URL found for '%s'", remote)
	}

	return gitRemoteToHTTPS(remoteURL), nil
}

func gitRemoteToHTTPS(remoteURL string) string {
	// Handle SSH short format: git@github.com:owner/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		remoteURL = strings.TrimPrefix(remoteURL, "git@")
		remoteURL = strings.Replace(remoteURL, ":", "/", 1)
		remoteURL = "https://" + remoteURL
	}

	// Handle SSH full format: ssh://git@github.com/owner/repo.git
	if strings.HasPrefix(remoteURL, "ssh://") {
		remoteURL = strings.TrimPrefix(remoteURL, "ssh://")
		// Remove user@ prefix if present (e.g., git@)
		if idx := strings.Index(remoteURL, "@"); idx != -1 {
			remoteURL = remoteURL[idx+1:]
		}
		remoteURL = "https://" + remoteURL
	}

	// Handle git:// protocol
	if strings.HasPrefix(remoteURL, "git://") {
		remoteURL = strings.Replace(remoteURL, "git://", "https://", 1)
	}

	// Remove .git suffix
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	return remoteURL
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

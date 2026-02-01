// Package notify provides OS notifications for session state transitions.
package notify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/gigurra/tofu/cmd/claude/common/config"
)

// stateDir returns the directory for notification state files.
func stateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".tofu", "notify-state")
}

// stateFile returns the path to a session's notification state file.
func stateFile(sessionID string) string {
	return filepath.Join(stateDir(), sessionID)
}

// OnStateTransition is called when a session changes state.
// It checks cooldown via file modification time and sends notification if appropriate.
// convTitle is optional - pass empty string if not available.
func OnStateTransition(sessionID, from, to, cwd, convTitle string) {
	cfg, err := config.Load()
	if err != nil || cfg.Notifications == nil || !cfg.Notifications.Enabled {
		return
	}

	// Check if transition matches config
	if !cfg.Notifications.MatchesTransition(from, to) {
		return
	}

	// Check cooldown via file modification time
	cooldown := time.Duration(cfg.Notifications.CooldownSeconds) * time.Second
	statePath := stateFile(sessionID)

	if info, err := os.Stat(statePath); err == nil {
		if time.Since(info.ModTime()) < cooldown {
			return
		}
	}

	// Send notification
	send(sessionID, to, cwd, convTitle)

	// Update state file (touch it)
	if err := os.MkdirAll(stateDir(), 0755); err == nil {
		// Create or update the file
		f, err := os.Create(statePath)
		if err == nil {
			f.Close()
		}
	}
}

// send actually sends the notification.
func send(sessionID, to, cwd, convTitle string) {
	// Build notification content
	projectName := filepath.Base(cwd)
	if projectName == "" || projectName == "." {
		projectName = "unknown"
	}

	statusDisplay := formatStatus(to)
	title := fmt.Sprintf("Claude: %s", statusDisplay)

	// Build body: first line is ID | Project, second line is conversation title
	var body string
	if convTitle != "" {
		body = fmt.Sprintf("%s | %s\n%s", shortID(sessionID), projectName, convTitle)
	} else {
		body = fmt.Sprintf("%s | %s", shortID(sessionID), projectName)
	}

	// Try beeep first (works on native Linux/macOS/Windows)
	err := beeep.Notify(title, body, "")
	if err != nil && isWSL() {
		// Fallback to PowerShell for WSL
		err = notifyWSL(title, body)
	}

	if err != nil {
		// Fallback to stderr (though this may not be visible in hooks)
		fmt.Fprintf(os.Stderr, "[notify] %s: %s\n", title, body)
	}
}

// isWSL detects if we're running in Windows Subsystem for Linux.
func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// notifyWSL sends a Windows Toast notification via PowerShell from WSL.
func notifyWSL(title, body string) error {
	// Escape single quotes for PowerShell
	title = strings.ReplaceAll(title, "'", "''")
	body = strings.ReplaceAll(body, "'", "''")

	// PowerShell script for Windows Toast notification
	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
$template = @'
<toast>
  <visual>
    <binding template="ToastText02">
      <text id="1">%s</text>
      <text id="2">%s</text>
    </binding>
  </visual>
</toast>
'@
$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Tofu').Show($toast)
`, title, body)

	cmd := exec.Command("/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe",
		"-NoProfile", "-NonInteractive", "-Command", script)
	return cmd.Run()
}

// formatStatus returns a human-readable status string.
func formatStatus(status string) string {
	switch status {
	case "working":
		return "Working"
	case "idle":
		return "Idle"
	case "awaiting_permission":
		return "Awaiting permission"
	case "awaiting_input":
		return "Awaiting input"
	case "exited":
		return "Exited"
	default:
		return status
	}
}

// shortID returns a shortened session ID for display.
func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// IsEnabled returns whether notifications are enabled.
func IsEnabled() bool {
	cfg, err := config.Load()
	if err != nil {
		return false
	}
	return cfg.Notifications != nil && cfg.Notifications.Enabled
}

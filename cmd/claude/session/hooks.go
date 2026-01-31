package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ClaudeSettings represents the structure of ~/.claude/settings.json
type ClaudeSettings struct {
	Hooks map[string][]HookMatcher `json:"hooks,omitempty"`
	// Preserve other fields
	Other map[string]json.RawMessage `json:"-"`
}

// HookMatcher represents a hook matcher configuration
type HookMatcher struct {
	Matcher string       `json:"matcher,omitempty"`
	Hooks   []HookConfig `json:"hooks"`
}

// HookConfig represents a single hook configuration
type HookConfig struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// TofuHookMarker is used to identify tofu-managed hooks
const TofuHookMarker = "tofu claude session status-callback"

// RequiredHooks defines the hooks tofu needs for status tracking
var RequiredHooks = map[string][]HookMatcher{
	"UserPromptSubmit": {
		{
			Hooks: []HookConfig{
				{Type: "command", Command: "tofu claude session status-callback working"},
			},
		},
	},
	"Stop": {
		{
			Hooks: []HookConfig{
				{Type: "command", Command: "tofu claude session status-callback idle"},
			},
		},
	},
	"Notification": {
		{
			Matcher: "permission_prompt",
			Hooks: []HookConfig{
				{Type: "command", Command: "tofu claude session status-callback awaiting_permission"},
			},
		},
		{
			Matcher: "elicitation_dialog",
			Hooks: []HookConfig{
				{Type: "command", Command: "tofu claude session status-callback awaiting_input"},
			},
		},
	},
}

// ClaudeSettingsPath returns the path to ~/.claude/settings.json
func ClaudeSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// CheckHooksInstalled checks if tofu hooks are installed in Claude settings
func CheckHooksInstalled() (installed bool, missing []string) {
	settingsPath := ClaudeSettingsPath()
	if settingsPath == "" {
		return false, []string{"all"}
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, []string{"all (settings.json not found)"}
		}
		return false, []string{"all (cannot read settings.json)"}
	}

	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return false, []string{"all (invalid settings.json)"}
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		return false, []string{"all (no hooks section)"}
	}

	var hooks map[string][]HookMatcher
	if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
		return false, []string{"all (invalid hooks section)"}
	}

	// Check each required hook event
	for event := range RequiredHooks {
		if !hasTofuHook(hooks, event) {
			missing = append(missing, event)
		}
	}

	return len(missing) == 0, missing
}

// hasTofuHook checks if a hook event has a tofu status-callback hook
func hasTofuHook(hooks map[string][]HookMatcher, event string) bool {
	matchers, ok := hooks[event]
	if !ok {
		return false
	}

	for _, matcher := range matchers {
		for _, hook := range matcher.Hooks {
			if hook.Type == "command" && containsTofuCallback(hook.Command) {
				return true
			}
		}
	}
	return false
}

func containsTofuCallback(cmd string) bool {
	return len(cmd) >= len(TofuHookMarker) && cmd[:len(TofuHookMarker)] == TofuHookMarker ||
		// Also check if it contains the marker anywhere
		findSubstring(cmd, TofuHookMarker)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// InstallHooks adds tofu hooks to Claude settings, preserving existing config
func InstallHooks() error {
	settingsPath := ClaudeSettingsPath()
	if settingsPath == "" {
		return fmt.Errorf("cannot determine Claude settings path")
	}

	// Ensure .claude directory exists
	claudeDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Read existing settings or start with empty object
	var settings map[string]json.RawMessage
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read settings: %w", err)
		}
		settings = make(map[string]json.RawMessage)
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse settings: %w", err)
		}
	}

	// Get existing hooks or create new
	var hooks map[string][]HookMatcher
	if hooksRaw, ok := settings["hooks"]; ok {
		if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
			return fmt.Errorf("failed to parse hooks: %w", err)
		}
	} else {
		hooks = make(map[string][]HookMatcher)
	}

	// Add missing tofu hooks
	for event, requiredMatchers := range RequiredHooks {
		if !hasTofuHook(hooks, event) {
			// Append our hooks to existing ones for this event
			hooks[event] = append(hooks[event], requiredMatchers...)
		}
	}

	// Serialize hooks back
	hooksData, err := json.Marshal(hooks)
	if err != nil {
		return fmt.Errorf("failed to serialize hooks: %w", err)
	}
	settings["hooks"] = hooksData

	// Write back settings with pretty formatting
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// EnsureHooksInstalled checks and optionally installs hooks, returning true if ready
func EnsureHooksInstalled(autoInstall bool, stdout, stderr *os.File) bool {
	installed, missing := CheckHooksInstalled()
	if installed {
		return true
	}

	if !autoInstall {
		fmt.Fprintf(stderr, "Warning: Tofu session hooks not installed in Claude settings.\n")
		fmt.Fprintf(stderr, "Missing hooks for: %v\n", missing)
		fmt.Fprintf(stderr, "Status tracking will not work. Install with: tofu claude session install-hooks\n\n")
		return false
	}

	fmt.Fprintf(stdout, "Installing tofu session hooks...\n")
	if err := InstallHooks(); err != nil {
		fmt.Fprintf(stderr, "Failed to install hooks: %v\n", err)
		return false
	}
	fmt.Fprintf(stdout, "Hooks installed successfully.\n\n")
	return true
}

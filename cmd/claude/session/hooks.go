package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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
const TofuHookMarker = "tofu claude session hook-callback"

// TofuHookCommand is the unified callback command for all hooks
const TofuHookCommand = "tofu claude session hook-callback"

// RequiredHooks defines the hooks tofu needs for status tracking
// All hooks use the same unified callback - it reads stdin and figures out what to do
var RequiredHooks = map[string][]HookMatcher{
	"UserPromptSubmit": {
		{
			Hooks: []HookConfig{
				{Type: "command", Command: TofuHookCommand},
			},
		},
	},
	"Stop": {
		{
			Hooks: []HookConfig{
				{Type: "command", Command: TofuHookCommand},
			},
		},
	},
	"PermissionRequest": {
		{
			Hooks: []HookConfig{
				{Type: "command", Command: TofuHookCommand},
			},
		},
	},
	"PostToolUse": {
		{
			Hooks: []HookConfig{
				{Type: "command", Command: TofuHookCommand},
			},
		},
	},
	"PostToolUseFailure": {
		{
			Hooks: []HookConfig{
				{Type: "command", Command: TofuHookCommand},
			},
		},
	},
	"Notification": {
		{
			// No matcher = catch all notifications, let tofu decide what to do
			Hooks: []HookConfig{
				{Type: "command", Command: TofuHookCommand},
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

	// Parse only what we need to check
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return false, []string{"all (invalid settings.json)"}
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		return false, []string{"all (no hooks section)"}
	}

	var hooks map[string]json.RawMessage
	if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
		return false, []string{"all (invalid hooks section)"}
	}

	// Check each required hook event by looking for our marker in the raw JSON
	for event := range RequiredHooks {
		eventHooks, ok := hooks[event]
		if !ok || !containsTofuCallback(string(eventHooks)) {
			missing = append(missing, event)
		}
	}

	return len(missing) == 0, missing
}

func containsTofuCallback(s string) bool {
	return findSubstring(s, TofuHookMarker)
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
// Uses map[string]json.RawMessage at each level to avoid touching data we don't need to modify
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

	// Get existing hooks as raw JSON map (preserves structure of events we don't touch)
	var hooks map[string]json.RawMessage
	if hooksRaw, ok := settings["hooks"]; ok {
		if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
			return fmt.Errorf("failed to parse hooks: %w", err)
		}
	} else {
		hooks = make(map[string]json.RawMessage)
	}

	// Add missing tofu hooks - only modify events that need our hooks
	for event, requiredMatchers := range RequiredHooks {
		eventHooksRaw, exists := hooks[event]

		// Check if this event already has our callback
		if exists && containsTofuCallback(string(eventHooksRaw)) {
			continue // Already installed, don't touch
		}

		if exists {
			// Event exists but doesn't have our hook - parse, append, re-serialize just this event
			var existingMatchers []json.RawMessage
			if err := json.Unmarshal(eventHooksRaw, &existingMatchers); err != nil {
				return fmt.Errorf("failed to parse %s hooks: %w", event, err)
			}

			// Append our matchers as raw JSON
			for _, matcher := range requiredMatchers {
				matcherJSON, err := json.Marshal(matcher)
				if err != nil {
					return fmt.Errorf("failed to serialize matcher: %w", err)
				}
				existingMatchers = append(existingMatchers, matcherJSON)
			}

			// Re-serialize just this event
			newEventHooks, err := json.Marshal(existingMatchers)
			if err != nil {
				return fmt.Errorf("failed to serialize %s hooks: %w", event, err)
			}
			hooks[event] = newEventHooks
		} else {
			// Event doesn't exist - create it with our matchers
			eventHooks, err := json.Marshal(requiredMatchers)
			if err != nil {
				return fmt.Errorf("failed to serialize %s hooks: %w", event, err)
			}
			hooks[event] = eventHooks
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

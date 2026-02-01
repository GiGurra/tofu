package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gigurra/tofu/cmd/claude/common/convindex"
	"github.com/gigurra/tofu/cmd/claude/common/notify"
	"github.com/spf13/cobra"
)

// HookCallbackInput represents the JSON input from any Claude Code hook
type HookCallbackInput struct {
	SessionID        string `json:"session_id"`
	TranscriptPath   string `json:"transcript_path"`
	Cwd              string `json:"cwd"`
	PermissionMode   string `json:"permission_mode,omitempty"`
	HookEventName    string `json:"hook_event_name"`
	NotificationType string `json:"notification_type,omitempty"`
	Message          string `json:"message,omitempty"`
	Prompt           string `json:"prompt,omitempty"`
	StopHookActive   bool   `json:"stop_hook_active,omitempty"`
	ToolName         string `json:"tool_name,omitempty"`
	AgentType        string `json:"agent_type,omitempty"`
	AgentID          string `json:"agent_id,omitempty"`
}

func HookCallbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "hook-callback",
		Short:  "Handle Claude Code hooks (internal)",
		Long:   "Unified callback for all Claude Code hooks. Reads hook data from stdin and updates session state accordingly.",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runHookCallback(); err != nil {
				// Silent failure - don't disrupt Claude's flow
				os.Exit(1)
			}
		},
	}
	return cmd
}

func runHookCallback() error {
	// Read hook input from stdin
	stdinData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var input HookCallbackInput
	if len(stdinData) > 0 {
		if err := json.Unmarshal(stdinData, &input); err != nil {
			return fmt.Errorf("failed to parse hook input: %w", err)
		}
	}

	// Log for debugging if TOFU_HOOK_DEBUG=true
	if os.Getenv("TOFU_HOOK_DEBUG") == "true" {
		_ = EnsureSessionsDir()
		debugFile, _ := os.OpenFile(DebugLogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if debugFile != nil {
			fmt.Fprintf(debugFile, "--- %s ---\n", time.Now().Format(time.RFC3339))
			fmt.Fprintf(debugFile, "Event: %s\n", input.HookEventName)
			if input.NotificationType != "" {
				fmt.Fprintf(debugFile, "NotificationType: %s\n", input.NotificationType)
			}
			if input.Message != "" {
				fmt.Fprintf(debugFile, "Message: %s\n", input.Message)
			}
			fmt.Fprintf(debugFile, "Raw: %s\n", string(stdinData))
			debugFile.Close()
		}
	}

	// Determine status based on hook event
	var newStatus string
	var statusDetail string

	switch input.HookEventName {
	case "UserPromptSubmit":
		newStatus = StatusWorking
		statusDetail = "UserPromptSubmit"

	case "PreToolUse":
		// Tool is about to execute
		newStatus = StatusWorking
		statusDetail = input.ToolName

	case "PostToolUse", "PostToolUseFailure":
		// Tool completed (success or failure) - back to working
		newStatus = StatusWorking
		statusDetail = input.ToolName

	case "SubagentStart", "SubagentStop":
		// Just log, don't update status (can fire after Stop and overwrite idle)
		return nil

	case "Stop":
		newStatus = StatusIdle
		statusDetail = ""

	case "PermissionRequest":
		newStatus = StatusAwaitingPermission
		statusDetail = input.ToolName
		if statusDetail == "" {
			statusDetail = "permission"
		}

	case "Notification":
		// Check notification type for legacy support
		switch input.NotificationType {
		case "permission_prompt":
			newStatus = StatusAwaitingPermission
			statusDetail = input.Message
		case "elicitation_dialog":
			newStatus = StatusAwaitingInput
			statusDetail = input.Message
		default:
			// Unknown notification type - log but don't update status
			return nil
		}

	default:
		// Unknown hook event - log but don't update status
		return nil
	}

	// Get or create session state
	state, err := getOrCreateSessionState(input)
	if err != nil || state == nil {
		return err
	}

	// Capture previous status for notification
	prevStatus := state.Status

	// Update status
	state.Status = newStatus
	state.StatusDetail = statusDetail
	state.Updated = time.Now()

	// Update ConvID from hook input if we don't have it yet
	if state.ConvID == "" && input.SessionID != "" {
		state.ConvID = input.SessionID
	}

	// Update PID if stale
	if state.PID > 0 && !IsProcessAlive(state.PID) {
		if newPID := FindClaudePID(); newPID > 0 {
			state.PID = newPID
		}
	} else if state.PID == 0 {
		if newPID := FindClaudePID(); newPID > 0 {
			state.PID = newPID
		}
	}

	// Save updated state
	if err := SaveSessionState(state); err != nil {
		return err
	}

	// Look up conversation title for notification
	convTitle := getConvTitle(state.ConvID, state.Cwd)

	// Notify on state transition (handles cooldown internally)
	notify.OnStateTransition(state.ID, prevStatus, newStatus, state.Cwd, convTitle)

	return nil
}

// getConvTitle looks up the conversation title from Claude's session index.
func getConvTitle(convID, cwd string) string {
	return convindex.GetConvTitle(convID, cwd)
}

// getOrCreateSessionState finds existing session or creates a new one
func getOrCreateSessionState(input HookCallbackInput) (*SessionState, error) {
	// Check for TOFU_SESSION_ID env var (session started via tofu)
	tofuSessionID := os.Getenv("TOFU_SESSION_ID")

	if tofuSessionID != "" {
		// Load existing session
		return LoadSessionState(tofuSessionID)
	}

	// Session wasn't started via tofu - try to auto-register
	if input.SessionID == "" {
		return nil, nil
	}

	// Check if we already have a session for this Claude conversation
	state := findSessionByConvID(input.SessionID)
	if state != nil {
		return state, nil
	}

	// Create a new auto-registered session
	return autoRegisterSessionFromHook(input), nil
}

// autoRegisterSessionFromHook creates a new session state for a Claude session
// that wasn't started via tofu
func autoRegisterSessionFromHook(input HookCallbackInput) *SessionState {
	// Find Claude's PID by walking up the process tree
	claudePID := FindClaudePID()
	if claudePID == 0 {
		return nil
	}

	// Check if we're inside tmux
	tmuxSession := GetCurrentTmuxSession()

	// Generate a session ID (first 8 chars of Claude's session ID)
	sessionID := input.SessionID
	if len(sessionID) > 8 {
		sessionID = sessionID[:8]
	}

	// Determine cwd
	cwd := input.Cwd
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	state := &SessionState{
		ID:          sessionID,
		TmuxSession: tmuxSession,
		PID:         claudePID,
		Cwd:         cwd,
		ConvID:      input.SessionID,
		Status:      StatusWorking,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	// Ensure sessions directory exists
	if err := EnsureSessionsDir(); err != nil {
		return nil
	}

	// Handle ID collision
	existingPath := SessionStatePath(sessionID)
	if _, err := os.Stat(existingPath); err == nil {
		existing, err := LoadSessionState(sessionID)
		if err == nil && existing.ConvID == input.SessionID {
			return existing
		}
		for i := 1; i < 100; i++ {
			newID := fmt.Sprintf("%s-%d", sessionID, i)
			if _, err := os.Stat(SessionStatePath(newID)); os.IsNotExist(err) {
				state.ID = newID
				break
			}
		}
	}

	// Save the new session
	if err := SaveSessionState(state); err != nil {
		return nil
	}

	// Write marker file
	markerPath := filepath.Join(SessionsDir(), state.ID+".auto")
	os.WriteFile(markerPath, []byte("auto-registered"), 0644)

	return state
}

package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type StatusCallbackParams struct {
	Status string `pos:"true" help:"New status (working, idle, awaiting_permission, awaiting_input)"`
}

// Valid status values for callbacks
const (
	StatusWorking            = "working"
	StatusIdle               = "idle"
	StatusAwaitingPermission = "awaiting_permission"
	StatusAwaitingInput      = "awaiting_input"
)

func StatusCallbackCmd() *cobra.Command {
	cmd := boa.CmdT[StatusCallbackParams]{
		Use:         "status-callback <status>",
		Short:       "Update session status (called by Claude hooks)",
		Long:        "Internal command called by Claude Code hooks to update session status. Not intended for direct use.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *StatusCallbackParams, cmd *cobra.Command, args []string) {
			if err := runStatusCallback(params); err != nil {
				// Silent failure - don't disrupt Claude's flow
				os.Exit(1)
			}
		},
	}.ToCobra()
	cmd.Hidden = true // Hide from help since it's for hooks
	return cmd
}

// HookInput represents the JSON input from Claude Code hooks
type HookInput struct {
	SessionID        string `json:"session_id"`
	Cwd              string `json:"cwd"`
	HookEventName    string `json:"hook_event_name"`
	NotificationType string `json:"notification_type,omitempty"`
}

func runStatusCallback(params *StatusCallbackParams) error {
	// Get tofu session ID from environment
	// If not set, silently succeed - this session wasn't started via tofu
	tofuSessionID := os.Getenv("TOFU_SESSION_ID")
	if tofuSessionID == "" {
		return nil
	}

	// Validate status
	switch params.Status {
	case StatusWorking, StatusIdle, StatusAwaitingPermission, StatusAwaitingInput:
		// Valid
	default:
		return fmt.Errorf("invalid status: %s", params.Status)
	}

	// Read hook input from stdin (optional, for logging/debugging)
	var hookInput HookInput
	stdinData, err := io.ReadAll(os.Stdin)
	if err == nil && len(stdinData) > 0 {
		json.Unmarshal(stdinData, &hookInput)
	}

	// Load existing session state
	state, err := LoadSessionState(tofuSessionID)
	if err != nil {
		return fmt.Errorf("failed to load session state: %w", err)
	}

	// Update status
	state.Status = params.Status
	state.Updated = time.Now()

	// Update ConvID from hook input if we don't have it yet
	// This happens when a session is started fresh (not resuming)
	if state.ConvID == "" && hookInput.SessionID != "" {
		state.ConvID = hookInput.SessionID
	}

	// Add detail if available
	if hookInput.HookEventName != "" {
		state.StatusDetail = hookInput.HookEventName
	}

	// Save updated state
	if err := SaveSessionState(state); err != nil {
		return fmt.Errorf("failed to save session state: %w", err)
	}

	return nil
}

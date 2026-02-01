package session

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/claude/common/inbox"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type MsgParams struct {
	SessionID string `pos:"true" help:"Session ID to message"`
	MsgType   string `short:"t" help:"Message type (focus)" default:"focus"`
}

func MsgCmd() *cobra.Command {
	return boa.CmdT[MsgParams]{
		Use:         "msg <session-id>",
		Short:       "Send a message to a session's inbox",
		Long:        "Posts a message to a session's inbox. The session will process it on its next hook callback.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *MsgParams, cmd *cobra.Command, args []string) {
			if params.SessionID == "" {
				fmt.Fprintf(os.Stderr, "Error: session ID required\n")
				os.Exit(1)
			}

			// Resolve partial session ID
			sessionID := ResolveSessionID(params.SessionID)
			if sessionID == "" {
				fmt.Fprintf(os.Stderr, "Error: session not found: %s\n", params.SessionID)
				os.Exit(1)
			}

			if err := inbox.Post(sessionID, params.MsgType, nil); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Posted %s message to session %s\n", params.MsgType, sessionID)
		},
	}.ToCobra()
}

// ResolveSessionID resolves a partial session ID to a full one.
// If an exact match exists, returns it. Otherwise looks for prefix matches.
func ResolveSessionID(partialID string) string {
	// Try exact match first
	if _, err := LoadSessionState(partialID); err == nil {
		return partialID
	}

	// Try prefix match
	states, err := ListSessionStates()
	if err != nil {
		return ""
	}

	var matches []string
	for _, state := range states {
		if len(state.ID) >= len(partialID) && state.ID[:len(partialID)] == partialID {
			matches = append(matches, state.ID)
		}
	}

	if len(matches) == 1 {
		return matches[0]
	}

	return "" // Ambiguous or not found
}

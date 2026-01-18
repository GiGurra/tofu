package conv

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type DeleteParams struct {
	ConvID string `pos:"true" help:"Conversation ID to delete"`
	Yes    bool   `short:"y" help:"Skip confirmation prompt"`
}

func DeleteCmd() *cobra.Command {
	return boa.CmdT[DeleteParams]{
		Use:         "delete",
		Aliases:     []string{"rm"},
		Short:       "Delete a Claude conversation",
		Long:        "Delete a Claude Code conversation from the current directory.",
		ParamEnrich: common.DefaultParamEnricher(),
		ValidArgsFunc: func(p *DeleteParams, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return getConversationCompletions(false), cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunFunc: func(params *DeleteParams, cmd *cobra.Command, args []string) {
			exitCode := RunDelete(params, os.Stdout, os.Stderr, os.Stdin)
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}.ToCobra()
}

func RunDelete(params *DeleteParams, stdout, stderr *os.File, stdin *os.File) int {
	// Extract just the ID from autocomplete format (e.g., "0459cd73_[tofu_claude]_prompt..." -> "0459cd73")
	convID := params.ConvID
	if idx := strings.Index(convID, "_"); idx > 0 {
		convID = convID[:idx]
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "Error getting current directory: %v\n", err)
		return 1
	}

	projectPath := GetClaudeProjectPath(cwd)

	// Load index
	index, err := LoadSessionsIndex(projectPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error loading sessions index: %v\n", err)
		return 1
	}

	// Find the conversation
	entry, _ := FindSessionByID(index, convID)
	if entry == nil {
		fmt.Fprintf(stderr, "Conversation %s not found in %s\n", convID, cwd)
		return 1
	}

	// Use the full session ID from the found entry
	fullID := entry.SessionID

	// Show what we're about to delete
	displayName := entry.CustomTitle
	if displayName == "" {
		displayName = truncatePrompt(entry.FirstPrompt, 50)
	}
	fmt.Fprintf(stdout, "Conversation: %s\n", fullID[:8])
	fmt.Fprintf(stdout, "Title/Prompt: %s\n", displayName)
	fmt.Fprintf(stdout, "Messages: %d\n", entry.MessageCount)

	// Confirm deletion
	if !params.Yes {
		fmt.Fprintf(stdout, "\nDelete this conversation? [y/N]: ")
		reader := bufio.NewReader(stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintf(stdout, "Aborted.\n")
			return 0
		}
	}

	// Delete conversation file
	convFile := filepath.Join(projectPath, fullID+".jsonl")
	if err := os.Remove(convFile); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(stderr, "Error deleting conversation file: %v\n", err)
		return 1
	}

	// Delete conversation directory if it exists
	convDir := filepath.Join(projectPath, fullID)
	if info, err := os.Stat(convDir); err == nil && info.IsDir() {
		if err := os.RemoveAll(convDir); err != nil {
			fmt.Fprintf(stderr, "Error deleting conversation directory: %v\n", err)
			return 1
		}
	}

	// Remove from index
	RemoveSessionByID(index, fullID)

	// Save index
	if err := SaveSessionsIndex(projectPath, index); err != nil {
		fmt.Fprintf(stderr, "Error saving sessions index: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Deleted conversation %s\n", fullID[:8])
	return 0
}

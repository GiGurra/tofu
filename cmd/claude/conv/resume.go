package conv

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type ResumeParams struct {
	ConvID string `pos:"true" help:"Conversation ID to resume (can be short prefix)"`
	Global bool   `short:"g" help:"Search for conversation across all projects"`
}

func ResumeCmd() *cobra.Command {
	return boa.CmdT[ResumeParams]{
		Use:         "resume",
		Short:       "Resume a Claude Code conversation",
		Long:        "Resume a Claude Code conversation by ID. Finds the conversation, changes to its project directory, and launches claude --resume.",
		ParamEnrich: common.DefaultParamEnricher(),
		ValidArgsFunc: func(p *ResumeParams, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			// Check if -g flag is set (p.Global may not be populated during completion)
			global, _ := cmd.Flags().GetBool("global")
			return getConversationCompletions(global), cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoFileComp
		},
		RunFunc: func(params *ResumeParams, cmd *cobra.Command, args []string) {
			exitCode := RunResume(params, os.Stdout, os.Stderr)
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}.ToCobra()
}

func getConversationCompletions(global bool) []string {
	var entries []SessionEntry

	if global {
		projectsDir := ClaudeProjectsDir()
		dirEntries, err := os.ReadDir(projectsDir)
		if err != nil {
			return nil
		}

		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() {
				continue
			}
			projPath := projectsDir + "/" + dirEntry.Name()
			index, err := LoadSessionsIndex(projPath)
			if err != nil {
				continue
			}
			entries = append(entries, index.Entries...)
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil
		}

		claudeProjectPath := GetClaudeProjectPath(cwd)
		index, err := LoadSessionsIndex(claudeProjectPath)
		if err != nil {
			return nil
		}

		entries = index.Entries
	}

	// Sort by modified date descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Modified > entries[j].Modified
	})

	// Format completions
	results := make([]string, len(entries))
	for i, e := range entries {
		results[i] = formatCompletion(e)
	}

	return results
}

func formatCompletion(e SessionEntry) string {
	// Format: <ID>_[name_]prompt__cr_datetime__u_datetime
	sanitize := func(s string) string {
		s = strings.ReplaceAll(s, "\t", "__")
		s = strings.ReplaceAll(s, " ", "_")
		s = strings.ReplaceAll(s, "\n", "_")
		s = strings.ReplaceAll(s, "\r", "")
		return s
	}

	id := e.SessionID[:8]

	// Use HasTitle/DisplayTitle for backwards compatible title display
	var namePart string
	if e.HasTitle() {
		namePart = "[" + sanitize(e.DisplayTitle()) + "]_"
	}

	prompt := sanitize(e.FirstPrompt)
	if len(prompt) > 40 {
		prompt = prompt[:37] + "..."
	}

	created := formatDateShort(e.Created)
	modified := formatDateShort(e.Modified)

	return fmt.Sprintf("%s_%s%s__cr_%s__u_%s", id, namePart, prompt, created, modified)
}

func formatDateShort(isoDate string) string {
	if len(isoDate) < 16 {
		return isoDate
	}
	// Extract YYYY-MM-DD_HH:MM from ISO format
	return strings.ReplaceAll(isoDate[:16], "T", "_")
}

func RunResume(params *ResumeParams, stdout, stderr *os.File) int {
	var entry *SessionEntry
	var projectPath string

	// Extract just the ID from autocomplete format (e.g., "0459cd73_[tofu_claude]_prompt..." -> "0459cd73")
	convID := params.ConvID
	if idx := strings.Index(convID, "_"); idx > 0 {
		convID = convID[:idx]
	}

	if params.Global {
		// Search all projects
		projectsDir := ClaudeProjectsDir()
		entries, err := os.ReadDir(projectsDir)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading projects directory: %v\n", err)
			return 1
		}

		for _, dirEntry := range entries {
			if !dirEntry.IsDir() {
				continue
			}
			projPath := projectsDir + "/" + dirEntry.Name()
			index, err := LoadSessionsIndex(projPath)
			if err != nil {
				continue
			}
			if found, _ := FindSessionByID(index, convID); found != nil {
				entry = found
				projectPath = found.ProjectPath
				break
			}
		}
	} else {
		// Search current directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "Error getting current directory: %v\n", err)
			return 1
		}

		claudeProjectPath := GetClaudeProjectPath(cwd)
		index, err := LoadSessionsIndex(claudeProjectPath)
		if err != nil {
			fmt.Fprintf(stderr, "Error loading sessions index: %v\n", err)
			return 1
		}

		entry, _ = FindSessionByID(index, convID)
		if entry != nil {
			projectPath = entry.ProjectPath
		}
	}

	if entry == nil {
		fmt.Fprintf(stderr, "Conversation %s not found\n", convID)
		if !params.Global {
			fmt.Fprintf(stderr, "Hint: use -g to search all projects\n")
		}
		return 1
	}

	// Show what we're doing
	displayName := entry.DisplayTitle()
	if len(displayName) > 50 {
		displayName = displayName[:47] + "..."
	}
	fmt.Fprintf(stdout, "Resuming [%s] in %s\n\n", displayName, projectPath)

	// Run claude --resume as a subprocess with connected I/O
	cmd := exec.Command("claude", "--resume", entry.SessionID)
	cmd.Dir = projectPath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(stderr, "Error running claude: %v\n", err)
		return 1
	}

	return 0
}

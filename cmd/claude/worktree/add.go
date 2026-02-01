package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	clcommon "github.com/gigurra/tofu/cmd/claude/common"
	"github.com/gigurra/tofu/cmd/claude/conv"
	"github.com/gigurra/tofu/cmd/claude/session"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type AddParams struct {
	Branch     string `pos:"true" help:"Branch name for the new worktree"`
	FromBranch string `long:"from-branch" optional:"true" help:"Base branch to create from (defaults to main/master)"`
	FromConv   string `long:"from-conv" optional:"true" help:"Conversation ID to copy to the new worktree"`
	Path       string `long:"path" optional:"true" help:"Custom path for the worktree (defaults to ../<repo>-<branch>)"`
	Detached   bool   `long:"detached" short:"d" help:"Don't start/attach to a Claude session"`
	Global     bool   `short:"g" help:"Search for conversation across all projects (with --from-conv)"`
}

func AddCmd() *cobra.Command {
	cmd := boa.CmdT[AddParams]{
		Use:         "add",
		Short:       "Create a new git worktree with a Claude session",
		Long:        "Create a new git worktree for parallel development.\n\nThis creates a new branch (if needed), sets up a worktree, optionally copies a conversation, and starts a Claude session.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *AddParams, cmd *cobra.Command, args []string) {
			if err := runAdd(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()

	// Register completions
	cmd.RegisterFlagCompletionFunc("from-branch", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return GetBranchCompletions(), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("from-conv", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		global, _ := cmd.Flags().GetBool("global")
		return clcommon.GetConversationCompletions(global), cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runAdd(params *AddParams) error {
	if params.Branch == "" {
		return fmt.Errorf("branch name is required")
	}

	// Get git info
	gitInfo, err := GetGitInfo()
	if err != nil {
		return err
	}

	// Determine base branch
	baseBranch := params.FromBranch
	if baseBranch == "" {
		baseBranch, err = GetDefaultBranch()
		if err != nil {
			return fmt.Errorf("could not determine base branch: %w (use --from-branch to specify)", err)
		}
	}

	// Check base branch exists
	if !BranchExists(baseBranch) {
		return fmt.Errorf("base branch %q does not exist", baseBranch)
	}

	// Determine worktree path
	worktreePath := params.Path
	if worktreePath == "" {
		// Default: ../<repo>-<branch>
		parentDir := filepath.Dir(gitInfo.RepoRoot)
		worktreePath = filepath.Join(parentDir, gitInfo.RepoName+"-"+params.Branch)
	}

	// Make path absolute
	if !filepath.IsAbs(worktreePath) {
		cwd, _ := os.Getwd()
		worktreePath = filepath.Join(cwd, worktreePath)
	}

	// Check if path already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("path already exists: %s", worktreePath)
	}

	// Check if branch already exists
	branchExists := BranchExists(params.Branch)

	fmt.Printf("Creating worktree...\n")
	fmt.Printf("  Branch: %s", params.Branch)
	if !branchExists {
		fmt.Printf(" (new, from %s)", baseBranch)
	}
	fmt.Printf("\n")
	fmt.Printf("  Path: %s\n", worktreePath)

	// Create worktree
	var gitCmd *exec.Cmd
	if branchExists {
		// Use existing branch
		gitCmd = exec.Command("git", "worktree", "add", worktreePath, params.Branch)
	} else {
		// Create new branch from base
		gitCmd = exec.Command("git", "worktree", "add", "-b", params.Branch, worktreePath, baseBranch)
	}

	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	fmt.Printf("\nWorktree created successfully.\n")

	// Copy conversation if specified
	var copiedConvID string
	if params.FromConv != "" {
		convID := clcommon.ExtractIDFromCompletion(params.FromConv)
		fmt.Printf("\nCopying conversation %s...\n", convID)

		copiedConvID, err = copyConversationToWorktree(convID, worktreePath, params.Global)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to copy conversation: %v\n", err)
		} else {
			fmt.Printf("  Copied to new project: %s\n", copiedConvID[:8])
		}
	}

	// Start session unless detached
	if !params.Detached {
		fmt.Printf("\nStarting Claude session...\n")

		newParams := &session.NewParams{
			Dir: worktreePath,
		}

		// If we copied a conversation, resume it
		if copiedConvID != "" {
			newParams.Resume = copiedConvID
		}

		return session.RunNew(newParams)
	}

	fmt.Printf("\nTo start a session:\n")
	fmt.Printf("  cd %s && tofu claude\n", worktreePath)
	if copiedConvID != "" {
		fmt.Printf("  # Or resume the copied conversation:\n")
		fmt.Printf("  cd %s && tofu claude --resume %s\n", worktreePath, copiedConvID[:8])
	}

	return nil
}

// copyConversationToWorktree copies a conversation to a new worktree's project directory
func copyConversationToWorktree(convID, worktreePath string, global bool) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	var srcEntry *conv.SessionEntry
	var srcProjectPath string
	dstProjectPath := conv.GetClaudeProjectPath(worktreePath)

	if global {
		// Search all projects
		projectsDir := conv.ClaudeProjectsDir()
		entries, err := os.ReadDir(projectsDir)
		if err != nil {
			return "", fmt.Errorf("failed to read projects directory: %w", err)
		}

		for _, dirEntry := range entries {
			if !dirEntry.IsDir() {
				continue
			}
			projPath := projectsDir + "/" + dirEntry.Name()
			index, err := conv.LoadSessionsIndex(projPath)
			if err != nil {
				continue
			}
			if found, _ := conv.FindSessionByID(index, convID); found != nil {
				srcEntry = found
				srcProjectPath = projPath
				break
			}
		}
	} else {
		srcProjectPath = conv.GetClaudeProjectPath(cwd)
		srcIndex, err := conv.LoadSessionsIndex(srcProjectPath)
		if err != nil {
			return "", fmt.Errorf("failed to load sessions index: %w", err)
		}
		srcEntry, _ = conv.FindSessionByID(srcIndex, convID)
	}

	if srcEntry == nil {
		if global {
			return "", fmt.Errorf("conversation %s not found", convID)
		}
		return "", fmt.Errorf("conversation %s not found in current project (use -g for global search)", convID)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(dstProjectPath, 0700); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Load or create destination index
	dstIndex, err := conv.LoadSessionsIndex(dstProjectPath)
	if err != nil {
		// Create new index
		dstIndex = &conv.SessionsIndex{Version: 1, Entries: []conv.SessionEntry{}}
	}

	// Generate new UUID
	newConvID := generateUUID()
	oldConvID := srcEntry.SessionID

	// Copy conversation file
	srcConvFile := filepath.Join(srcProjectPath, oldConvID+".jsonl")
	dstConvFile := filepath.Join(dstProjectPath, newConvID+".jsonl")

	if err := conv.CopyConversationFile(srcConvFile, dstConvFile, oldConvID, newConvID); err != nil {
		return "", fmt.Errorf("failed to copy conversation file: %w", err)
	}

	// Copy conversation directory if exists
	srcConvDir := filepath.Join(srcProjectPath, oldConvID)
	dstConvDir := filepath.Join(dstProjectPath, newConvID)
	if info, err := os.Stat(srcConvDir); err == nil && info.IsDir() {
		if err := conv.CopyDir(srcConvDir, dstConvDir); err != nil {
			return "", fmt.Errorf("failed to copy conversation directory: %w", err)
		}
	}

	// Get file info for new entry
	dstInfo, err := os.Stat(dstConvFile)
	if err != nil {
		return "", fmt.Errorf("failed to stat destination file: %w", err)
	}

	// Create new entry
	now := formatTime()
	newEntry := conv.SessionEntry{
		SessionID:    newConvID,
		FullPath:     dstConvFile,
		FileMtime:    dstInfo.ModTime().UnixMilli(),
		FirstPrompt:  srcEntry.FirstPrompt,
		Summary:      srcEntry.Summary,
		CustomTitle:  srcEntry.CustomTitle,
		MessageCount: srcEntry.MessageCount,
		Created:      now,
		Modified:     now,
		GitBranch:    srcEntry.GitBranch,
		ProjectPath:  worktreePath,
		IsSidechain:  srcEntry.IsSidechain,
	}

	dstIndex.Entries = append(dstIndex.Entries, newEntry)

	if err := conv.SaveSessionsIndex(dstProjectPath, dstIndex); err != nil {
		return "", fmt.Errorf("failed to save destination index: %w", err)
	}

	return newConvID, nil
}

// generateUUID generates a new UUID v4
func generateUUID() string {
	return uuid.New().String()
}

// formatTime returns current time in RFC3339 format (local time)
func formatTime() string {
	return time.Now().Format(time.RFC3339)
}

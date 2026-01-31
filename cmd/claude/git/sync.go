package git

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/claude/conv"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type SyncParams struct {
	KeepLocal  bool `long:"keep-local" help:"On conflict, keep local version without asking"`
	KeepRemote bool `long:"keep-remote" help:"On conflict, keep remote version without asking"`
	DryRun     bool `long:"dry-run" help:"Show what would be synced without making changes"`
}

func SyncCmd() *cobra.Command {
	return boa.CmdT[SyncParams]{
		Use:         "sync",
		Short:       "Sync Claude conversations with remote",
		Long:        "Sync local Claude conversations with the remote git repository.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *SyncParams, cmd *cobra.Command, args []string) {
			if err := runSync(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runSync(params *SyncParams) error {
	syncDir := SyncDir()
	projectsDir := ProjectsDir()

	if !IsInitialized() {
		return fmt.Errorf("git sync not initialized. Run 'tofu claude git init <repo-url>' first")
	}

	fmt.Printf("Syncing Claude conversations...\n\n")

	// Step 1: Pull remote changes into sync dir
	fmt.Printf("Step 1: Pulling remote changes...\n")
	if !params.DryRun {
		pullCmd := exec.Command("git", "pull", "--ff-only")
		pullCmd.Dir = syncDir
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			// Might be first sync or diverged - try fetch instead
			fmt.Printf("  (pull failed, trying reset to remote)\n")
			fetchCmd := exec.Command("git", "fetch", "origin")
			fetchCmd.Dir = syncDir
			fetchCmd.Run()

			// Reset to remote if we have uncommitted local changes in sync dir
			// (this is safe because local source of truth is ~/.claude/projects)
			remoteBranch := getRemoteBranch(syncDir)
			if remoteBranch != "" {
				resetCmd := exec.Command("git", "reset", "--hard", "origin/"+remoteBranch)
				resetCmd.Dir = syncDir
				resetCmd.Run()
			}
		}
	}

	// Step 2: Merge local conversations into sync directory
	fmt.Printf("\nStep 2: Merging local conversations...\n")
	if err := mergeLocalToSync(projectsDir, syncDir, params); err != nil {
		return fmt.Errorf("failed to merge local changes: %w", err)
	}

	// Step 3: Commit changes
	fmt.Printf("\nStep 3: Committing changes...\n")
	if !params.DryRun {
		// Add all changes
		addCmd := exec.Command("git", "add", "-A")
		addCmd.Dir = syncDir
		if output, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add failed: %s\n%w", output, err)
		}

		// Check if there are changes to commit
		diffCmd := exec.Command("git", "diff", "--cached", "--quiet")
		diffCmd.Dir = syncDir
		if err := diffCmd.Run(); err != nil {
			// There are changes to commit
			hostname, _ := os.Hostname()
			commitMsg := fmt.Sprintf("sync from %s at %s", hostname, time.Now().Format(time.RFC3339))
			commitCmd := exec.Command("git", "commit", "-m", commitMsg)
			commitCmd.Dir = syncDir
			if output, err := commitCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("git commit failed: %s\n%w", output, err)
			}
			fmt.Printf("  Committed changes\n")
		} else {
			fmt.Printf("  No changes to commit\n")
		}
	}

	// Step 4: Push to remote
	fmt.Printf("\nStep 4: Pushing to remote...\n")
	if !params.DryRun {
		pushCmd := exec.Command("git", "push", "-u", "origin", "HEAD")
		pushCmd.Dir = syncDir
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			return fmt.Errorf("git push failed: %w", err)
		}
	}

	// Step 5: Copy merged results back to local
	fmt.Printf("\nStep 5: Updating local conversations...\n")
	if !params.DryRun {
		if err := copySyncToProjects(syncDir, projectsDir); err != nil {
			return fmt.Errorf("failed to update local projects: %w", err)
		}
	}

	fmt.Printf("\nSync complete!\n")
	return nil
}

// getRemoteBranch returns the remote branch name (main or master), or empty if none
func getRemoteBranch(syncDir string) string {
	// Try main first
	checkCmd := exec.Command("git", "rev-parse", "--verify", "origin/main")
	checkCmd.Dir = syncDir
	if err := checkCmd.Run(); err == nil {
		return "main"
	}
	// Try master
	checkCmd2 := exec.Command("git", "rev-parse", "--verify", "origin/master")
	checkCmd2.Dir = syncDir
	if err := checkCmd2.Run(); err == nil {
		return "master"
	}
	return ""
}

// mergeLocalToSync merges local conversations into the sync directory
// This is an intelligent merge: for sessions-index.json we merge entries,
// for conversation files we copy if newer or prompt on conflict
func mergeLocalToSync(projectsDir, syncDir string, params *SyncParams) error {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No projects yet
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		localProject := filepath.Join(projectsDir, entry.Name())
		syncProject := filepath.Join(syncDir, entry.Name())

		if params.DryRun {
			fmt.Printf("  Would merge: %s\n", entry.Name())
			continue
		}

		// Create project directory in sync if needed
		if err := os.MkdirAll(syncProject, 0755); err != nil {
			return err
		}

		// Merge this project
		if err := mergeProject(localProject, syncProject, params); err != nil {
			fmt.Printf("  Warning: merge issue in %s: %v\n", entry.Name(), err)
		}
	}

	return nil
}

// copyProjectsToSync copies conversation files from projects to sync directory
func copyProjectsToSync(projectsDir, syncDir string, dryRun bool) error {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No projects yet
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		srcProject := filepath.Join(projectsDir, entry.Name())
		dstProject := filepath.Join(syncDir, entry.Name())

		if dryRun {
			fmt.Printf("  Would copy: %s\n", entry.Name())
			continue
		}

		// Create project directory in sync
		if err := os.MkdirAll(dstProject, 0755); err != nil {
			return err
		}

		// Copy all files (conversations and index)
		files, err := os.ReadDir(srcProject)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				// Copy subdirectories (conversation data dirs)
				if err := conv.CopyDir(filepath.Join(srcProject, f.Name()), filepath.Join(dstProject, f.Name())); err != nil {
					fmt.Printf("  Warning: failed to copy %s: %v\n", f.Name(), err)
				}
			} else {
				// Copy files
				if err := conv.CopyFile(filepath.Join(srcProject, f.Name()), filepath.Join(dstProject, f.Name())); err != nil {
					fmt.Printf("  Warning: failed to copy %s: %v\n", f.Name(), err)
				}
			}
		}
	}

	return nil
}

// copySyncToProjects copies merged results back to projects directory
func copySyncToProjects(syncDir, projectsDir string) error {
	entries, err := os.ReadDir(syncDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" {
			continue
		}

		srcProject := filepath.Join(syncDir, entry.Name())
		dstProject := filepath.Join(projectsDir, entry.Name())

		// Create project directory
		if err := os.MkdirAll(dstProject, 0755); err != nil {
			return err
		}

		// Copy all files
		files, err := os.ReadDir(srcProject)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				if err := conv.CopyDir(filepath.Join(srcProject, f.Name()), filepath.Join(dstProject, f.Name())); err != nil {
					fmt.Printf("  Warning: failed to copy %s: %v\n", f.Name(), err)
				}
			} else {
				if err := conv.CopyFile(filepath.Join(srcProject, f.Name()), filepath.Join(dstProject, f.Name())); err != nil {
					fmt.Printf("  Warning: failed to copy %s: %v\n", f.Name(), err)
				}
			}
		}
	}

	return nil
}

// mergeProject merges a source project into destination project
// Used for merging local conversations into sync directory
func mergeProject(srcProject, dstProject string, params *SyncParams) error {
	// Ensure dst project exists
	os.MkdirAll(dstProject, 0755)

	// Merge sessions-index.json
	if err := mergeSessionsIndex(srcProject, dstProject); err != nil {
		fmt.Printf("    Warning: index merge issue: %v\n", err)
	}

	// Merge conversation files from source
	srcFiles, _ := os.ReadDir(srcProject)
	for _, f := range srcFiles {
		if f.Name() == "sessions-index.json" {
			continue // Already handled
		}

		srcPath := filepath.Join(srcProject, f.Name())
		dstPath := filepath.Join(dstProject, f.Name())

		if f.IsDir() {
			// Copy directory (subagents, etc)
			if err := conv.CopyDir(srcPath, dstPath); err != nil {
				fmt.Printf("    Warning: failed to copy dir %s: %v\n", f.Name(), err)
			}
			continue
		}

		// Check if file exists in destination
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			// Source only - copy it
			conv.CopyFile(srcPath, dstPath)
			continue
		}

		// Both exist - check if different
		if filesEqual(srcPath, dstPath) {
			continue // Same content
		}

		// Different content - for now, source (local) wins
		// This is safe because we pulled remote first, so sync has remote state
		// and local changes should overwrite
		if strings.HasSuffix(f.Name(), ".jsonl") {
			if err := handleConversationConflict(srcPath, dstPath, f.Name(), params); err != nil {
				fmt.Printf("    Warning: conflict resolution failed for %s: %v\n", f.Name(), err)
			}
		}
	}

	return nil
}

// mergeSessionsIndex intelligently merges src sessions-index.json into dst
// src = user's local projects, dst = sync dir (which has remote state after pull)
func mergeSessionsIndex(srcProject, dstProject string) error {
	srcIndexPath := filepath.Join(srcProject, "sessions-index.json")
	dstIndexPath := filepath.Join(dstProject, "sessions-index.json")

	// Load src (local) index
	var srcIndex conv.SessionsIndex
	if data, err := os.ReadFile(srcIndexPath); err == nil {
		json.Unmarshal(data, &srcIndex)
	}

	// Load dst (sync/remote) index
	var dstIndex conv.SessionsIndex
	if data, err := os.ReadFile(dstIndexPath); err == nil {
		json.Unmarshal(data, &dstIndex)
	} else {
		dstIndex = conv.SessionsIndex{Version: 1}
	}

	// Merge: union of entries, prefer newer modified timestamp
	merged := make(map[string]conv.SessionEntry)

	// Add all dst (remote) entries first
	for _, e := range dstIndex.Entries {
		merged[e.SessionID] = e
	}

	// Merge src (local) entries - local wins on conflict if newer
	for _, e := range srcIndex.Entries {
		if existing, ok := merged[e.SessionID]; ok {
			// Both have this entry - keep the one with newer Modified timestamp
			existingTime, _ := time.Parse(time.RFC3339, existing.Modified)
			srcTime, _ := time.Parse(time.RFC3339, e.Modified)
			if srcTime.After(existingTime) {
				merged[e.SessionID] = e
			}
		} else {
			// Local only - add it
			merged[e.SessionID] = e
		}
	}

	// Build result
	dstIndex.Entries = make([]conv.SessionEntry, 0, len(merged))
	for _, e := range merged {
		dstIndex.Entries = append(dstIndex.Entries, e)
	}

	// Save merged index to dst
	data, err := json.MarshalIndent(dstIndex, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dstIndexPath, data, 0644)
}

// handleConversationConflict handles a conflict in a conversation file
// srcPath = user's local file, dstPath = sync dir file (has remote state)
func handleConversationConflict(srcPath, dstPath, filename string, params *SyncParams) error {
	if params.KeepLocal {
		fmt.Printf("    Conflict in %s: keeping local (--keep-local)\n", filename)
		return conv.CopyFile(srcPath, dstPath)
	}

	if params.KeepRemote {
		fmt.Printf("    Conflict in %s: keeping remote (--keep-remote)\n", filename)
		return nil // dst already has remote state
	}

	// Count messages in each
	localLines := countLines(srcPath)
	remoteLines := countLines(dstPath)

	fmt.Printf("\n    Conflict: %s\n", filename)
	fmt.Printf("      Local:  %d messages\n", localLines)
	fmt.Printf("      Remote: %d messages\n", remoteLines)
	fmt.Printf("    Keep which version? [l]ocal / [r]emote / [s]kip: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "l", "local":
		fmt.Printf("    Keeping local version\n")
		return conv.CopyFile(srcPath, dstPath)
	case "r", "remote":
		fmt.Printf("    Keeping remote version\n")
		return nil // dst already has remote state
	default:
		fmt.Printf("    Skipping (keeping remote)\n")
		return nil
	}
}

// filesEqual checks if two files have the same content
func filesEqual(path1, path2 string) bool {
	data1, err1 := os.ReadFile(path1)
	data2, err2 := os.ReadFile(path2)
	if err1 != nil || err2 != nil {
		return false
	}
	if len(data1) != len(data2) {
		return false
	}
	for i := range data1 {
		if data1[i] != data2[i] {
			return false
		}
	}
	return true
}

// countLines counts lines in a file (approximate message count for JSONL)
func countLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}

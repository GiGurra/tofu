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

	// Step 1: Copy local projects to sync directory
	fmt.Printf("Step 1: Copying local conversations to sync directory...\n")
	if err := copyProjectsToSync(projectsDir, syncDir, params.DryRun); err != nil {
		return fmt.Errorf("failed to copy projects: %w", err)
	}

	// Step 2: Fetch remote changes
	fmt.Printf("\nStep 2: Fetching remote changes...\n")
	if !params.DryRun {
		fetchCmd := exec.Command("git", "fetch", "origin")
		fetchCmd.Dir = syncDir
		fetchCmd.Stdout = os.Stdout
		fetchCmd.Stderr = os.Stderr
		if err := fetchCmd.Run(); err != nil {
			// Might be first push, continue
			fmt.Printf("  (no remote changes or first sync)\n")
		}
	}

	// Step 3: Check for remote branch and merge
	fmt.Printf("\nStep 3: Merging remote changes...\n")
	if !params.DryRun {
		// Check if remote branch exists
		checkCmd := exec.Command("git", "rev-parse", "--verify", "origin/main")
		checkCmd.Dir = syncDir
		if err := checkCmd.Run(); err != nil {
			// Also try master
			checkCmd2 := exec.Command("git", "rev-parse", "--verify", "origin/master")
			checkCmd2.Dir = syncDir
			if err := checkCmd2.Run(); err != nil {
				fmt.Printf("  (no remote branch yet - first sync)\n")
			}
		} else {
			// Remote exists, do the merge
			if err := mergeRemoteChanges(syncDir, params); err != nil {
				return fmt.Errorf("merge failed: %w", err)
			}
		}
	}

	// Step 4: Commit local changes
	fmt.Printf("\nStep 4: Committing changes...\n")
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

	// Step 5: Push to remote
	fmt.Printf("\nStep 5: Pushing to remote...\n")
	if !params.DryRun {
		// Try to push, set upstream if needed
		pushCmd := exec.Command("git", "push", "-u", "origin", "HEAD")
		pushCmd.Dir = syncDir
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			return fmt.Errorf("git push failed: %w", err)
		}
	}

	// Step 6: Copy merged results back to projects
	fmt.Printf("\nStep 6: Updating local conversations...\n")
	if !params.DryRun {
		if err := copySyncToProjects(syncDir, projectsDir); err != nil {
			return fmt.Errorf("failed to update local projects: %w", err)
		}
	}

	fmt.Printf("\nSync complete!\n")
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

// mergeRemoteChanges fetches and merges remote changes
func mergeRemoteChanges(syncDir string, params *SyncParams) error {
	// Create a temp directory for the remote version
	tempDir, err := os.MkdirTemp("", "tofu-claude-merge-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Clone the remote into temp
	cloneCmd := exec.Command("git", "clone", "--depth=1", syncDir, tempDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone for merge: %s\n%w", output, err)
	}

	// Fetch latest from origin in temp
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Dir = tempDir
	fetchCmd.Run()

	// Reset to origin/main (or master)
	resetCmd := exec.Command("git", "reset", "--hard", "origin/main")
	resetCmd.Dir = tempDir
	if err := resetCmd.Run(); err != nil {
		resetCmd2 := exec.Command("git", "reset", "--hard", "origin/master")
		resetCmd2.Dir = tempDir
		resetCmd2.Run()
	}

	// Now merge the remote (tempDir) into our sync directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" {
			continue
		}

		remoteProject := filepath.Join(tempDir, entry.Name())
		localProject := filepath.Join(syncDir, entry.Name())

		// Merge this project
		if err := mergeProject(remoteProject, localProject, params); err != nil {
			fmt.Printf("  Warning: merge issue in %s: %v\n", entry.Name(), err)
		}
	}

	return nil
}

// mergeProject merges a single project directory
func mergeProject(remoteProject, localProject string, params *SyncParams) error {
	// Ensure local project exists
	os.MkdirAll(localProject, 0755)

	// Merge sessions-index.json
	if err := mergeSessionsIndex(remoteProject, localProject); err != nil {
		fmt.Printf("    Warning: index merge issue: %v\n", err)
	}

	// Merge conversation files
	remoteFiles, _ := os.ReadDir(remoteProject)
	for _, f := range remoteFiles {
		if f.Name() == "sessions-index.json" {
			continue // Already handled
		}

		remotePath := filepath.Join(remoteProject, f.Name())
		localPath := filepath.Join(localProject, f.Name())

		if f.IsDir() {
			// Copy directory if not exists locally
			if _, err := os.Stat(localPath); os.IsNotExist(err) {
				conv.CopyDir(remotePath, localPath)
			}
			continue
		}

		// Check if file exists locally
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			// Remote only - copy it
			conv.CopyFile(remotePath, localPath)
			continue
		}

		// Both exist - check if different
		if filesEqual(remotePath, localPath) {
			continue // Same content
		}

		// Conflict! Handle based on flags or ask user
		if strings.HasSuffix(f.Name(), ".jsonl") {
			if err := handleConversationConflict(remotePath, localPath, f.Name(), params); err != nil {
				fmt.Printf("    Warning: conflict resolution failed for %s: %v\n", f.Name(), err)
			}
		}
	}

	return nil
}

// mergeSessionsIndex intelligently merges two sessions-index.json files
func mergeSessionsIndex(remoteProject, localProject string) error {
	remoteIndexPath := filepath.Join(remoteProject, "sessions-index.json")
	localIndexPath := filepath.Join(localProject, "sessions-index.json")

	// Load remote index
	var remoteIndex conv.SessionsIndex
	if data, err := os.ReadFile(remoteIndexPath); err == nil {
		json.Unmarshal(data, &remoteIndex)
	}

	// Load local index
	var localIndex conv.SessionsIndex
	if data, err := os.ReadFile(localIndexPath); err == nil {
		json.Unmarshal(data, &localIndex)
	} else {
		localIndex = conv.SessionsIndex{Version: 1}
	}

	// Merge: union of entries, prefer newer modified timestamp
	merged := make(map[string]conv.SessionEntry)

	// Add all local entries
	for _, e := range localIndex.Entries {
		merged[e.SessionID] = e
	}

	// Merge remote entries
	for _, e := range remoteIndex.Entries {
		if existing, ok := merged[e.SessionID]; ok {
			// Both have this entry - keep the one with newer Modified timestamp
			existingTime, _ := time.Parse(time.RFC3339, existing.Modified)
			remoteTime, _ := time.Parse(time.RFC3339, e.Modified)
			if remoteTime.After(existingTime) {
				merged[e.SessionID] = e
			}
		} else {
			// Remote only
			merged[e.SessionID] = e
		}
	}

	// Build result
	localIndex.Entries = make([]conv.SessionEntry, 0, len(merged))
	for _, e := range merged {
		localIndex.Entries = append(localIndex.Entries, e)
	}

	// Save merged index
	data, err := json.MarshalIndent(localIndex, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(localIndexPath, data, 0644)
}

// handleConversationConflict handles a conflict in a conversation file
func handleConversationConflict(remotePath, localPath, filename string, params *SyncParams) error {
	if params.KeepLocal {
		fmt.Printf("    Conflict in %s: keeping local (--keep-local)\n", filename)
		return nil
	}

	if params.KeepRemote {
		fmt.Printf("    Conflict in %s: keeping remote (--keep-remote)\n", filename)
		return conv.CopyFile(remotePath, localPath)
	}

	// Count messages in each
	localLines := countLines(localPath)
	remoteLines := countLines(remotePath)

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
		return nil
	case "r", "remote":
		fmt.Printf("    Keeping remote version\n")
		return conv.CopyFile(remotePath, localPath)
	default:
		fmt.Printf("    Skipping (keeping local)\n")
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

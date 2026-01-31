package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/GiGurra/cmder"
	"github.com/spf13/cobra"
)

type Params struct {
	Dir              string `pos:"true" optional:"true" help:"Parent directory containing git repos (defaults to current directory)" default:"."`
	Prune            bool   `short:"p" help:"Delete all local branches except the default branch"`
	DryRun           bool   `short:"n" name:"dry-run" help:"Show what would be done without making changes"`
	Pattern          string `name:"pattern" req:"false" help:"Only process directories matching this regex pattern"`
	Parallel         int    `short:"P" name:"parallel" help:"Number of repos to process in parallel" default:"10"`
	StashUncommitted bool   `name:"stash" help:"Stash uncommitted changes before syncing (pop after)"`
	DropUncommitted  bool   `name:"drop" help:"Discard uncommitted changes before syncing (DANGEROUS)"`
}

type RepoResult struct {
	Name           string
	Skipped        bool
	SkipReason     string
	Error          error
	PreviousBranch string
	DefaultBranch  string
	PullOutput     string
	NewCommits     int
	PrunedBranches []string
	Stashed        bool
	Dropped        bool
	// Post-sync state
	FinalBranch     string
	HasUncommitted  bool
	OnDefaultBranch bool
	OnFeatureBranch bool
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:   "sync [dir]",
		Short: "Sync git repo(s) to their default branch",
		Long: `If the specified directory is a git repository, syncs that repo.
Otherwise, iterates through all directories one level below and syncs each git repository.

For each git repository:
  - Checks for uncommitted changes
  - Switches to the default branch
  - Pulls the latest changes
  - Optionally prunes non-default local branches

Examples:
  tofu git sync                    # Sync current dir (if git repo) or all repos in subdirs
  tofu git sync ~/projects         # Sync all repos in ~/projects
  tofu git sync ~/projects/myrepo  # Sync single repo
  tofu git sync --prune            # Also delete non-default branches
  tofu git sync --dry-run          # Preview what would happen
  tofu git sync --parallel 1       # Process repos sequentially
  tofu git sync --stash            # Stash uncommitted changes, sync, then pop
  tofu git sync --pattern "^api-"  # Only sync repos matching pattern`,
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := run(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func run(params *Params) error {
	if params.StashUncommitted && params.DropUncommitted {
		return fmt.Errorf("--stash and --drop are mutually exclusive")
	}

	targetDir, err := filepath.Abs(params.Dir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory: %w", err)
	}

	// Check if target directory itself is a git repo
	if isGitRepo(targetDir) {
		return runSingleRepo(targetDir, params)
	}

	// Otherwise, look for git repos in subdirectories
	return runMultipleRepos(targetDir, params)
}

func runSingleRepo(repoPath string, params *Params) error {
	if params.DryRun {
		fmt.Println("DRY RUN - no changes will be made")
		fmt.Println()
	}

	repoName := filepath.Base(repoPath)
	result := processRepo(filepath.Dir(repoPath), repoName, params)
	printProgress(result, params.DryRun)
	printReport([]RepoResult{result}, params)
	return nil
}

func runMultipleRepos(parentDir string, params *Params) error {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var pattern *regexp.Regexp
	if params.Pattern != "" {
		pattern, err = regexp.Compile(params.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
	}

	// Collect directories to process
	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if pattern != nil && !pattern.MatchString(entry.Name()) {
			continue
		}
		repoPath := filepath.Join(parentDir, entry.Name())
		if isGitRepo(repoPath) {
			dirs = append(dirs, entry.Name())
		}
	}

	if len(dirs) == 0 {
		fmt.Println("No git repositories found")
		return nil
	}

	if params.DryRun {
		fmt.Println("DRY RUN - no changes will be made")
		fmt.Println()
	}

	results := make([]RepoResult, len(dirs))

	if params.Parallel <= 1 {
		// Sequential processing
		for i, dir := range dirs {
			results[i] = processRepo(parentDir, dir, params)
			printProgress(results[i], params.DryRun)
		}
	} else {
		// Parallel processing
		var wg sync.WaitGroup
		sem := make(chan struct{}, params.Parallel)
		var mu sync.Mutex

		for i, dir := range dirs {
			wg.Add(1)
			go func(idx int, dirName string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				result := processRepo(parentDir, dirName, params)

				mu.Lock()
				results[idx] = result
				printProgress(result, params.DryRun)
				mu.Unlock()
			}(i, dir)
		}
		wg.Wait()
	}

	printReport(results, params)
	return nil
}

func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func processRepo(parentDir, name string, params *Params) RepoResult {
	repoPath := filepath.Join(parentDir, name)
	result := RepoResult{Name: name}

	// Check for uncommitted changes
	hasChanges, err := hasUncommittedChanges(repoPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to check status: %w", err)
		return result
	}

	if hasChanges {
		if params.StashUncommitted {
			if !params.DryRun {
				if err := stashChanges(repoPath); err != nil {
					result.Error = fmt.Errorf("failed to stash: %w", err)
					return result
				}
			}
			result.Stashed = true
		} else if params.DropUncommitted {
			if !params.DryRun {
				if err := dropChanges(repoPath); err != nil {
					result.Error = fmt.Errorf("failed to drop changes: %w", err)
					return result
				}
			}
			result.Dropped = true
		} else {
			result.Skipped = true
			result.SkipReason = "uncommitted changes"
			return result
		}
	}

	// Get current branch
	currentBranch, err := getCurrentBranch(repoPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to get current branch: %w", err)
		return result
	}
	result.PreviousBranch = currentBranch

	// Get default branch
	defaultBranch, err := getDefaultBranch(repoPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to detect default branch: %w", err)
		return result
	}
	result.DefaultBranch = defaultBranch

	// Switch to default branch if needed
	if currentBranch != defaultBranch {
		if !params.DryRun {
			if err := checkoutBranch(repoPath, defaultBranch); err != nil {
				result.Error = fmt.Errorf("failed to checkout %s: %w", defaultBranch, err)
				return result
			}
		}
	}

	// Pull latest changes
	if !params.DryRun {
		output, commits, err := pullChanges(repoPath)
		if err != nil {
			result.Error = fmt.Errorf("failed to pull: %w", err)
			return result
		}
		result.PullOutput = output
		result.NewCommits = commits
	}

	// Prune branches if requested
	if params.Prune {
		branches, err := getNonDefaultBranches(repoPath, defaultBranch)
		if err != nil {
			result.Error = fmt.Errorf("failed to list branches: %w", err)
			return result
		}
		result.PrunedBranches = branches

		if !params.DryRun {
			for _, branch := range branches {
				if err := deleteBranch(repoPath, branch); err != nil {
					// Continue on error, just note it
					result.Error = fmt.Errorf("failed to delete branch %s: %w", branch, err)
				}
			}
		}
	}

	// Pop stash if we stashed earlier
	if result.Stashed && !params.DryRun {
		_ = popStash(repoPath) // Best effort, ignore errors
	}

	// Capture post-sync state
	capturePostSyncState(repoPath, &result)

	return result
}

func capturePostSyncState(repoPath string, result *RepoResult) {
	// Get current branch
	if branch, err := getCurrentBranch(repoPath); err == nil {
		result.FinalBranch = branch
		result.OnDefaultBranch = branch == result.DefaultBranch
		result.OnFeatureBranch = branch != result.DefaultBranch && branch != ""
	}

	// Check for uncommitted changes
	if hasChanges, err := hasUncommittedChanges(repoPath); err == nil {
		result.HasUncommitted = hasChanges
	}
}

func hasUncommittedChanges(repoPath string) (bool, error) {
	result := cmder.New("git", "-C", repoPath, "status", "--porcelain").
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())

	if result.Err != nil {
		return false, result.Err
	}

	return strings.TrimSpace(result.StdOut) != "", nil
}

func stashChanges(repoPath string) error {
	result := cmder.New("git", "-C", repoPath, "stash", "push", "-m", "tofu-git-sync-autostash").
		WithAttemptTimeout(30 * time.Second).
		Run(context.Background())
	return result.Err
}

func popStash(repoPath string) error {
	result := cmder.New("git", "-C", repoPath, "stash", "pop").
		WithAttemptTimeout(30 * time.Second).
		Run(context.Background())
	return result.Err
}

func dropChanges(repoPath string) error {
	// Reset staged and unstaged changes
	result := cmder.New("git", "-C", repoPath, "reset", "--hard", "HEAD").
		WithAttemptTimeout(30 * time.Second).
		Run(context.Background())
	if result.Err != nil {
		return result.Err
	}

	// Clean untracked files
	result = cmder.New("git", "-C", repoPath, "clean", "-fd").
		WithAttemptTimeout(30 * time.Second).
		Run(context.Background())
	return result.Err
}

func getCurrentBranch(repoPath string) (string, error) {
	result := cmder.New("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD").
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())

	if result.Err != nil {
		return "", result.Err
	}

	return strings.TrimSpace(result.StdOut), nil
}

func getDefaultBranch(repoPath string) (string, error) {
	// Try to get from origin/HEAD first
	result := cmder.New("git", "-C", repoPath, "symbolic-ref", "refs/remotes/origin/HEAD").
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())

	if result.Err == nil {
		ref := strings.TrimSpace(result.StdOut)
		// refs/remotes/origin/main -> main
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check common default branch names
	for _, name := range []string{"main", "master", "develop", "trunk"} {
		result := cmder.New("git", "-C", repoPath, "rev-parse", "--verify", name).
			WithAttemptTimeout(5 * time.Second).
			Run(context.Background())
		if result.Err == nil {
			return name, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

func checkoutBranch(repoPath, branch string) error {
	result := cmder.New("git", "-C", repoPath, "checkout", branch).
		WithAttemptTimeout(30 * time.Second).
		Run(context.Background())
	return result.Err
}

func pullChanges(repoPath string) (string, int, error) {
	// Get current commit before pull
	beforeResult := cmder.New("git", "-C", repoPath, "rev-parse", "HEAD").
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())
	if beforeResult.Err != nil {
		return "", 0, beforeResult.Err
	}
	beforeCommit := strings.TrimSpace(beforeResult.StdOut)

	// Pull
	pullResult := cmder.New("git", "-C", repoPath, "pull", "--ff-only").
		WithAttemptTimeout(120 * time.Second).
		Run(context.Background())
	if pullResult.Err != nil {
		return pullResult.StdOut + pullResult.StdErr, 0, pullResult.Err
	}

	// Get commit count
	afterResult := cmder.New("git", "-C", repoPath, "rev-parse", "HEAD").
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())
	if afterResult.Err != nil {
		return pullResult.StdOut, 0, nil
	}
	afterCommit := strings.TrimSpace(afterResult.StdOut)

	if beforeCommit == afterCommit {
		return pullResult.StdOut, 0, nil
	}

	// Count commits between before and after
	countResult := cmder.New("git", "-C", repoPath, "rev-list", "--count", beforeCommit+".."+afterCommit).
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())
	if countResult.Err != nil {
		return pullResult.StdOut, 0, nil
	}

	var count int
	fmt.Sscanf(strings.TrimSpace(countResult.StdOut), "%d", &count)
	return pullResult.StdOut, count, nil
}

func getNonDefaultBranches(repoPath, defaultBranch string) ([]string, error) {
	result := cmder.New("git", "-C", repoPath, "branch", "--format=%(refname:short)").
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())

	if result.Err != nil {
		return nil, result.Err
	}

	var branches []string
	for _, line := range strings.Split(result.StdOut, "\n") {
		branch := strings.TrimSpace(line)
		if branch != "" && branch != defaultBranch {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

func deleteBranch(repoPath, branch string) error {
	result := cmder.New("git", "-C", repoPath, "branch", "-D", branch).
		WithAttemptTimeout(10 * time.Second).
		Run(context.Background())
	return result.Err
}

func printProgress(result RepoResult, dryRun bool) {
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}

	fmt.Printf("%s%s/\n", prefix, result.Name)

	if result.Error != nil {
		fmt.Printf("  ! Error: %v\n", result.Error)
		return
	}

	if result.Skipped {
		fmt.Printf("  ~ Skipped: %s\n", result.SkipReason)
		return
	}

	if result.Stashed {
		fmt.Printf("  * Stashed changes\n")
	}
	if result.Dropped {
		fmt.Printf("  * Dropped uncommitted changes\n")
	}

	if result.NewCommits > 0 {
		fmt.Printf("  + Pulled %s (%d new commits)\n", result.DefaultBranch, result.NewCommits)
	} else {
		fmt.Printf("  = %s already up to date\n", result.DefaultBranch)
	}

	if len(result.PrunedBranches) > 0 {
		fmt.Printf("  - Pruned: %s\n", strings.Join(result.PrunedBranches, ", "))
	}
}

func printReport(results []RepoResult, params *Params) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("SYNC REPORT")
	fmt.Println(strings.Repeat("=", 50))

	var synced, skipped, errored int
	var totalNewCommits int
	var totalPruned int

	for _, r := range results {
		if r.Error != nil {
			errored++
		} else if r.Skipped {
			skipped++
		} else {
			synced++
			totalNewCommits += r.NewCommits
			totalPruned += len(r.PrunedBranches)
		}
	}

	fmt.Printf("\nRepositories: %d total\n", len(results))
	fmt.Printf("  Synced:  %d\n", synced)
	if skipped > 0 {
		fmt.Printf("  Skipped: %d\n", skipped)
	}
	if errored > 0 {
		fmt.Printf("  Errors:  %d\n", errored)
	}

	if totalNewCommits > 0 {
		fmt.Printf("\nNew commits pulled: %d\n", totalNewCommits)
	}
	if totalPruned > 0 {
		fmt.Printf("Branches pruned:    %d\n", totalPruned)
	}

	if params.DryRun {
		fmt.Println("\n(DRY RUN - no changes were made)")
	}

	// List errors if any
	if errored > 0 {
		fmt.Println("\nErrors:")
		for _, r := range results {
			if r.Error != nil {
				fmt.Printf("  %s: %v\n", r.Name, r.Error)
			}
		}
	}

	// List skipped repos
	if skipped > 0 {
		fmt.Println("\nSkipped:")
		for _, r := range results {
			if r.Skipped {
				fmt.Printf("  %s: %s\n", r.Name, r.SkipReason)
			}
		}
	}

	// Post-sync state summary
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("CURRENT STATE")
	fmt.Println(strings.Repeat("-", 50))

	// Repos on default branch
	var onDefault []string
	for _, r := range results {
		if r.OnDefaultBranch {
			onDefault = append(onDefault, r.Name)
		}
	}
	if len(onDefault) > 0 {
		fmt.Printf("\nOn default branch (%d):\n", len(onDefault))
		for _, name := range onDefault {
			fmt.Printf("  %s\n", name)
		}
	}

	// Repos on feature branches
	var onFeature []RepoResult
	for _, r := range results {
		if r.OnFeatureBranch {
			onFeature = append(onFeature, r)
		}
	}
	if len(onFeature) > 0 {
		fmt.Printf("\nOn feature branch (%d):\n", len(onFeature))
		for _, r := range onFeature {
			fmt.Printf("  %s: %s\n", r.Name, r.FinalBranch)
		}
	}

	// Repos with uncommitted changes
	var withChanges []string
	for _, r := range results {
		if r.HasUncommitted {
			withChanges = append(withChanges, r.Name)
		}
	}
	if len(withChanges) > 0 {
		fmt.Printf("\nWith uncommitted changes (%d):\n", len(withChanges))
		for _, name := range withChanges {
			fmt.Printf("  %s\n", name)
		}
	}

	if len(onDefault) > 0 && len(onFeature) == 0 && len(withChanges) == 0 {
		fmt.Println("\nAll repos clean and on default branch!")
	}
}

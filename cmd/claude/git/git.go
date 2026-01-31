package git

import (
	"os"
	"path/filepath"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "git",
		Short: "Sync Claude conversations across devices via git",
		Long: `Sync Claude conversations across multiple computers using a git repository.

This keeps ~/.claude/projects_sync as a git working directory separate from
the actual ~/.claude/projects, giving full control over the merge process.

Usage:
  tofu claude git init <repo-url>   # Set up sync with a remote repo
  tofu claude git sync              # Sync local and remote conversations
  tofu claude git status            # Show sync status`,
		SubCmds: []*cobra.Command{
			InitCmd(),
			SyncCmd(),
			StatusCmd(),
			RepairCmd(),
		},
	}.ToCobra()
}

// SyncDir returns the path to the sync directory (~/.claude/projects_sync)
func SyncDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects_sync")
}

// ProjectsDir returns the path to the actual projects directory (~/.claude/projects)
func ProjectsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// IsInitialized checks if the sync directory is a git repository
func IsInitialized() bool {
	gitDir := filepath.Join(SyncDir(), ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

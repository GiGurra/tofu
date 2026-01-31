package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/claude/conv"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type RepairParams struct {
	DryRun bool `long:"dry-run" help:"Show what would be repaired without making changes"`
}

func RepairCmd() *cobra.Command {
	return boa.CmdT[RepairParams]{
		Use:   "repair",
		Short: "Repair sync directory by canonicalizing project paths",
		Long: `Repair the sync directory by merging equivalent project directories.

Uses the path mappings from ~/.claude/sync_config.json to identify
project directories that should be merged (e.g., different home paths
across machines).

This command:
1. Scans the sync directory for project directories
2. Groups them by canonical path (using sync_config.json)
3. Merges equivalent directories together
4. Removes the non-canonical duplicates`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *RepairParams, cmd *cobra.Command, args []string) {
			if err := runRepair(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runRepair(params *RepairParams) error {
	syncDir := SyncDir()

	if !IsInitialized() {
		return fmt.Errorf("git sync not initialized. Run 'tofu claude git init <repo-url>' first")
	}

	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load sync config: %w", err)
	}

	if len(config.Homes) == 0 && len(config.Dirs) == 0 {
		fmt.Printf("No path mappings configured in %s\n", ConfigPath())
		fmt.Printf("Nothing to repair.\n")
		return nil
	}

	fmt.Printf("Scanning sync directory for projects to repair...\n")
	fmt.Printf("Config: homes=%v, dirs=%d groups\n\n", config.Homes, len(config.Dirs))

	// Find all project directories in sync
	entries, err := os.ReadDir(syncDir)
	if err != nil {
		return fmt.Errorf("failed to read sync directory: %w", err)
	}

	// Group projects by canonical name
	groups := make(map[string][]string) // canonical -> [original dirs]
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" {
			continue
		}

		canonical := config.CanonicalizeProjectDir(entry.Name())
		groups[canonical] = append(groups[canonical], entry.Name())
	}

	// Find groups that need merging (more than one entry, or entry differs from canonical)
	mergeCount := 0
	for canonical, originals := range groups {
		needsMerge := len(originals) > 1 || (len(originals) == 1 && originals[0] != canonical)

		if !needsMerge {
			continue
		}

		mergeCount++
		fmt.Printf("Merge group: %s\n", canonical)
		for _, orig := range originals {
			if orig == canonical {
				fmt.Printf("  - %s (canonical)\n", orig)
			} else {
				fmt.Printf("  - %s -> merge into canonical\n", orig)
			}
		}

		if params.DryRun {
			fmt.Printf("  [dry-run] Would merge %d directories\n\n", len(originals))
			continue
		}

		// Perform the merge
		canonicalPath := filepath.Join(syncDir, canonical)

		// Ensure canonical directory exists
		if err := os.MkdirAll(canonicalPath, 0755); err != nil {
			return fmt.Errorf("failed to create canonical directory: %w", err)
		}

		// Merge each non-canonical directory into canonical
		for _, orig := range originals {
			if orig == canonical {
				continue
			}

			origPath := filepath.Join(syncDir, orig)
			fmt.Printf("  Merging %s...\n", orig)

			if err := mergeProjectIntoCanonical(origPath, canonicalPath); err != nil {
				fmt.Printf("    Warning: merge failed: %v\n", err)
				continue
			}

			// Remove the original (now merged) directory
			if err := os.RemoveAll(origPath); err != nil {
				fmt.Printf("    Warning: failed to remove %s: %v\n", orig, err)
			}
		}
		fmt.Println()
	}

	if mergeCount == 0 {
		fmt.Printf("No projects need repair.\n")
	} else if params.DryRun {
		fmt.Printf("Would repair %d project groups. Run without --dry-run to apply.\n", mergeCount)
	} else {
		fmt.Printf("Repaired %d project groups.\n", mergeCount)
		fmt.Printf("\nRun 'tofu claude git sync' to commit and push the repairs.\n")
	}

	return nil
}

// mergeProjectIntoCanonical merges a source project into the canonical project
func mergeProjectIntoCanonical(srcPath, dstPath string) error {
	// Merge sessions-index.json
	srcIndex := filepath.Join(srcPath, "sessions-index.json")
	dstIndex := filepath.Join(dstPath, "sessions-index.json")

	if _, err := os.Stat(srcIndex); err == nil {
		if err := mergeSessionsIndexFiles(srcIndex, dstIndex); err != nil {
			fmt.Printf("    Warning: index merge issue: %v\n", err)
		}
	}

	// Copy all other files/directories
	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name() == "sessions-index.json" {
			continue // Already handled
		}

		srcFile := filepath.Join(srcPath, entry.Name())
		dstFile := filepath.Join(dstPath, entry.Name())

		// Check if destination exists
		if _, err := os.Stat(dstFile); err == nil {
			// Destination exists - skip (prefer existing canonical version)
			continue
		}

		// Copy to canonical location
		if entry.IsDir() {
			if err := conv.CopyDir(srcFile, dstFile); err != nil {
				fmt.Printf("    Warning: failed to copy dir %s: %v\n", entry.Name(), err)
			}
		} else {
			if err := conv.CopyFile(srcFile, dstFile); err != nil {
				fmt.Printf("    Warning: failed to copy file %s: %v\n", entry.Name(), err)
			}
		}
	}

	return nil
}

// mergeSessionsIndexFiles merges src sessions-index.json into dst
func mergeSessionsIndexFiles(srcPath, dstPath string) error {
	// Use the existing merge function with project dirs
	srcDir := filepath.Dir(srcPath)
	dstDir := filepath.Dir(dstPath)
	return mergeSessionsIndex(srcDir, dstDir)
}

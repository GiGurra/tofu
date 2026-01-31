package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// SyncConfig holds path mapping configuration for cross-machine sync
type SyncConfig struct {
	// Homes lists equivalent home directories across machines
	// The first entry is the canonical form
	// e.g., ["/home/gigur", "/Users/johkjo"]
	Homes []string `json:"homes"`

	// Dirs lists groups of equivalent directories
	// Each group's first entry is the canonical form
	// e.g., [["/home/gigur/git", "/Users/johkjo/git/personal"]]
	Dirs [][]string `json:"dirs"`
}

// ConfigPath returns the path to the sync config file
// Prefers sync dir (auto-synced) over local ~/.claude
func ConfigPath() string {
	// First check sync dir (shared across machines)
	syncPath := filepath.Join(SyncDir(), "sync_config.json")
	if _, err := os.Stat(syncPath); err == nil {
		return syncPath
	}

	// Fall back to local ~/.claude
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "sync_config.json")
}

// LoadConfig loads the sync config, returning empty config if not found
func LoadConfig() (*SyncConfig, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SyncConfig{}, nil
		}
		return nil, err
	}

	var config SyncConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveConfig saves the sync config
func SaveConfig(config *SyncConfig) error {
	path := ConfigPath()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ProjectDirToPath converts a Claude project dir name back to a path
// e.g., "-home-gigur-git-tofu" -> "/home/gigur/git/tofu"
func ProjectDirToPath(projectDir string) string {
	// Remove leading dash and convert dashes to path separators
	if strings.HasPrefix(projectDir, "-") {
		projectDir = projectDir[1:]
	}
	return "/" + strings.ReplaceAll(projectDir, "-", "/")
}

// PathToProjectDir converts a path to Claude project dir name
// e.g., "/home/gigur/git/tofu" -> "-home-gigur-git-tofu"
func PathToProjectDir(path string) string {
	// Normalize path separators and remove leading slash
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return "-" + strings.ReplaceAll(path, "/", "-")
}

// CanonicalizeProjectDir converts a project dir to its canonical form
// by applying path mappings from the config
//
// Order of operations:
// 1. Convert project dir to path
// 2. Apply dirs mappings (most specific first)
// 3. Apply homes mappings
// 4. Convert back to project dir
func (c *SyncConfig) CanonicalizeProjectDir(projectDir string) string {
	if c == nil {
		return projectDir
	}

	// Convert to path for easier manipulation
	path := ProjectDirToPath(projectDir)

	// Apply dirs mappings (check all entries in each group)
	path = c.applyDirsMapping(path)

	// Apply homes mappings
	path = c.applyHomesMapping(path)

	// Convert back to project dir
	return PathToProjectDir(path)
}

// applyDirsMapping applies directory mappings to a path
// If path starts with any dir in a group, replace with the canonical (first) dir
func (c *SyncConfig) applyDirsMapping(path string) string {
	for _, group := range c.Dirs {
		if len(group) < 2 {
			continue
		}
		canonical := group[0]
		for _, dir := range group[1:] {
			if strings.HasPrefix(path, dir) {
				// Replace this dir prefix with the canonical one
				return canonical + path[len(dir):]
			}
		}
	}
	return path
}

// applyHomesMapping applies home directory mappings to a path
// If path starts with any home, replace with the canonical (first) home
func (c *SyncConfig) applyHomesMapping(path string) string {
	if len(c.Homes) < 2 {
		return path
	}
	canonical := c.Homes[0]
	for _, home := range c.Homes[1:] {
		if strings.HasPrefix(path, home) {
			return canonical + path[len(home):]
		}
	}
	return path
}

// FindEquivalentProjectDirs returns all project dir names that map to the same canonical form
func (c *SyncConfig) FindEquivalentProjectDirs(projectDir string) []string {
	if c == nil {
		return []string{projectDir}
	}

	canonical := c.CanonicalizeProjectDir(projectDir)
	result := []string{canonical}

	// Generate all possible variants by applying reverse mappings
	canonicalPath := ProjectDirToPath(canonical)

	// For each home variant
	for _, home := range c.Homes {
		if len(c.Homes) > 0 && strings.HasPrefix(canonicalPath, c.Homes[0]) {
			variant := home + canonicalPath[len(c.Homes[0]):]
			variantDir := PathToProjectDir(variant)
			if variantDir != canonical && !contains(result, variantDir) {
				result = append(result, variantDir)
			}
		}
	}

	// For each dirs variant
	for _, group := range c.Dirs {
		if len(group) < 2 {
			continue
		}
		if strings.HasPrefix(canonicalPath, group[0]) {
			for _, dir := range group[1:] {
				variant := dir + canonicalPath[len(group[0]):]
				variantDir := PathToProjectDir(variant)
				if !contains(result, variantDir) {
					result = append(result, variantDir)
				}
			}
		}
	}

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// LocalizePath converts a canonical path to the local machine's equivalent path
// This is the reverse of canonicalization - it expands canonical paths to local paths
// localHome should be the current machine's home directory
func (c *SyncConfig) LocalizePath(path, localHome string) string {
	if c == nil || len(c.Homes) == 0 {
		return path
	}

	canonical := c.Homes[0]

	// Find which local home/dir mapping applies to this machine
	// by checking if localHome matches any of the configured homes
	var localPrefix string
	for _, home := range c.Homes {
		if home == localHome || strings.HasPrefix(localHome, home) {
			localPrefix = home
			break
		}
	}

	// If we couldn't find a matching local prefix, can't localize
	if localPrefix == "" || localPrefix == canonical {
		return path
	}

	// Check dirs mappings first (more specific)
	for _, group := range c.Dirs {
		if len(group) < 2 {
			continue
		}
		canonicalDir := group[0]
		if strings.HasPrefix(path, canonicalDir) {
			// Find local equivalent in this group
			for _, dir := range group[1:] {
				if strings.HasPrefix(dir, localPrefix) {
					path = dir + path[len(canonicalDir):]
					break
				}
			}
		}
	}

	// Apply homes mapping
	if strings.HasPrefix(path, canonical) {
		path = localPrefix + path[len(canonical):]
	}

	// Also localize embedded project directory names (e.g., in fullPath)
	path = c.localizeEmbeddedProjectDir(path, localHome)

	return path
}

// localizeEmbeddedProjectDir finds and localizes project dir names embedded in paths
// e.g., /home/local/.claude/projects/-home-canonical-git-myproject/session.jsonl
// becomes /home/local/.claude/projects/-home-local-projects-myproject/session.jsonl
func (c *SyncConfig) localizeEmbeddedProjectDir(path, localHome string) string {
	// Look for .claude/projects/ pattern
	projectsMarker := ".claude/projects/"
	idx := strings.Index(path, projectsMarker)
	if idx == -1 {
		return path
	}

	// Find the project dir name (starts after projects/, ends at next /)
	start := idx + len(projectsMarker)
	end := strings.Index(path[start:], "/")
	if end == -1 {
		// Project dir is the last component
		projectDir := path[start:]
		localDir := c.LocalizeProjectDir(projectDir, localHome)
		if projectDir != localDir {
			return path[:start] + localDir
		}
		return path
	}

	// Extract, localize, and replace
	projectDir := path[start : start+end]
	localDir := c.LocalizeProjectDir(projectDir, localHome)
	if projectDir != localDir {
		return path[:start] + localDir + path[start+end:]
	}
	return path
}

// LocalizeProjectDir converts a project dir to the local machine's form
// This is the reverse of CanonicalizeProjectDir
// e.g., "-home-gigur-git-tofu" -> "-Users-johkjo-git-personal-tofu" (on Mac)
func (c *SyncConfig) LocalizeProjectDir(projectDir, localHome string) string {
	if c == nil || len(c.Homes) == 0 {
		return projectDir
	}

	// Convert to path, localize, convert back
	path := ProjectDirToPath(projectDir)
	localPath := c.LocalizePath(path, localHome)
	return PathToProjectDir(localPath)
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

type WatchPatternType string

const (
	WatchPatternTypeRegex   WatchPatternType = "regex"
	WatchPatternTypeLiteral WatchPatternType = "literal"
	WatchPatternTypeGlob    WatchPatternType = "glob"
)

type WatchParams struct {
	Execute          string           `short:"e" required:"true" help:"Command to execute when files change."`
	Patterns         []string         `short:"p" optional:"true" help:"File patterns to watch (optional, watches all files if not specified)."`
	PatternType      WatchPatternType `optional:"true" help:"Type of pattern matching (regex, literal, glob)." default:"glob" alts:"regex,literal,glob"`
	Recursive        bool             `short:"r" optional:"true" help:"Watch directories recursively." default:"true"`
	IncludeHidden    bool             `optional:"true" help:"Include hidden files and directories." default:"false"`
	Exclude          []string         `optional:"true" help:"Patterns to exclude (glob style)."`
	PreviousProcess  string           `optional:"true" help:"Action for previous process (kill, wait)." default:"kill" alts:"kill,wait"`
	HandleShutdown   string           `optional:"true" help:"Action when process exits (restart, ignore)." default:"ignore" alts:"restart,ignore"`
	RestartPolicy    string           `optional:"true" help:"Restart policy (exponential-backoff)." default:"exponential-backoff" alts:"exponential-backoff"`
	MinBackoffMillis int64            `optional:"true" help:"Minimum backoff duration in milliseconds." default:"1000"`
	MaxBackoffMillis int64            `optional:"true" help:"Maximum backoff duration in milliseconds." default:"10000"`
	MaxRestarts      int              `optional:"true" help:"Maximum number of automatic restarts." default:"10"`
	Dirs             []string         `pos:"true" optional:"true" help:"Directories to watch (defaults to current directory)." default:"."`
}

type ProcessRunner interface {
	Start() error
	Wait() error
	Kill() error
}

type ProcessFactory func() ProcessRunner

type RealProcessRunner struct {
	cmd *exec.Cmd
}

func (p *RealProcessRunner) Start() error {
	return p.cmd.Start()
}

func (p *RealProcessRunner) Wait() error {
	return p.cmd.Wait()
}

func WatchCmd() *cobra.Command {
	return boa.CmdT[WatchParams]{
		Use:         "watch",
		Short:       "Watch files and execute a command on change",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *WatchParams, cmd *cobra.Command, args []string) {
			factory := NewProcessRunner(params)
			if err := runWatch(cmd.Context(), params, factory); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "watch: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runWatch(ctx context.Context, params *WatchParams, factory ProcessFactory) error {
	// Create fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	// Compile regex patterns if needed
	var regexPatterns []*regexp.Regexp
	if params.PatternType == WatchPatternTypeRegex {
		for _, p := range params.Patterns {
			r, err := regexp.Compile(p)
			if err != nil {
				return fmt.Errorf("error compiling regex '%s': %w", p, err)
			}
			regexPatterns = append(regexPatterns, r)
		}
	}

	// Compile exclude patterns
	var excludePatterns []*regexp.Regexp
	for _, p := range params.Exclude {
		// Convert glob to regex for consistent handling
		pattern := globToRegex(p)
		r, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("error compiling exclude pattern '%s': %w", p, err)
		}
		excludePatterns = append(excludePatterns, r)
	}

	// Check if path should be excluded
	isExcluded := func(path string) bool {
		for _, r := range excludePatterns {
			if r.MatchString(path) {
				return true
			}
		}
		return false
	}

	// Check if path matches watch patterns
	matches := func(path string) bool {
		if len(params.Patterns) == 0 {
			return true
		}

		name := filepath.Base(path)
		for i, p := range params.Patterns {
			switch params.PatternType {
			case WatchPatternTypeGlob:
				matched, _ := filepath.Match(p, name)
				if matched {
					return true
				}
			case WatchPatternTypeLiteral:
				if strings.Contains(name, p) {
					return true
				}
			case WatchPatternTypeRegex:
				if regexPatterns[i].MatchString(name) {
					return true
				}
			}
		}
		return false
	}

	// Add directories to watcher recursively
	watchedDirs := make(map[string]bool)
	var watchMutex sync.Mutex

	addWatch := func(path string) error {
		watchMutex.Lock()
		defer watchMutex.Unlock()

		if watchedDirs[path] {
			return nil
		}

		if isExcluded(path) {
			return nil
		}

		if err := watcher.Add(path); err != nil {
			return err
		}

		watchedDirs[path] = true
		return nil
	}

	removeWatch := func(path string) {
		watchMutex.Lock()
		defer watchMutex.Unlock()

		if watchedDirs[path] {
			_ = watcher.Remove(path)
			delete(watchedDirs, path)
		}
	}

	// Recursively add all directories
	addRecursive := func(root string) error {
		return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if d.IsDir() {
				// Skip hidden directories unless explicitly allowed
				if !params.IncludeHidden && strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
					return filepath.SkipDir
				}

				if isExcluded(path) {
					return filepath.SkipDir
				}

				if err := addWatch(path); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to watch %s: %v\n", path, err)
				}
			}

			return nil
		})
	}

	// Initialize watches for all specified directories
	dirs := params.Dirs
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	for _, dir := range dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", dir, err)
		}

		if params.Recursive {
			if err := addRecursive(absDir); err != nil {
				return fmt.Errorf("failed to add watches recursively for %s: %w", absDir, err)
			}
		} else {
			if err := addWatch(absDir); err != nil {
				return fmt.Errorf("failed to watch %s: %w", absDir, err)
			}
		}
	}

	fmt.Printf("Watching %d directories...\n", len(watchedDirs))

	// Channel for triggering command execution
	changeChan := make(chan string, 10)

	// Debounce rapid file changes
	var debounceTimer *time.Timer
	var debounceMutex sync.Mutex
	debounceDelay := 100 * time.Millisecond

	triggerChange := func(path string) {
		debounceMutex.Lock()
		defer debounceMutex.Unlock()

		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		debounceTimer = time.AfterFunc(debounceDelay, func() {
			select {
			case changeChan <- path:
			default:
			}
		})
	}

	// Watch for filesystem events
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Skip if file doesn't match patterns
				if !matches(event.Name) {
					continue
				}

				// Skip excluded paths
				if isExcluded(event.Name) {
					continue
				}

				// Handle directory creation for recursive watching
				if params.Recursive && (event.Op&fsnotify.Create == fsnotify.Create) {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						if err := addRecursive(event.Name); err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to watch new directory %s: %v\n", event.Name, err)
						}
					}
				}

				// Handle directory removal
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					removeWatch(event.Name)
				}

				// Trigger command execution on relevant events
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					triggerChange(event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				_, _ = fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
			}
		}
	}()

	// Process management
	var cmdProcess ProcessRunner
	var cmdMutex sync.Mutex
	processRunning := false
	processDone := make(chan error, 1)

	killProcess := func() {
		cmdMutex.Lock()
		defer cmdMutex.Unlock()
		if cmdProcess != nil && processRunning {
			_ = cmdProcess.Kill()
		}
	}

	startProcess := func() {
		cmdMutex.Lock()
		if processRunning {
			cmdMutex.Unlock()
			return
		}

		fmt.Printf("Running: %s\n", params.Execute)
		cmd := factory()

		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start command: %v\n", err)
			cmdMutex.Unlock()
			processDone <- err
			return
		}

		cmdProcess = cmd
		processRunning = true
		cmdMutex.Unlock()

		go func() {
			err := cmd.Wait()
			cmdMutex.Lock()
			processRunning = false
			cmdMutex.Unlock()
			processDone <- err
		}()
	}

	// Handle system signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// State management
	restartCount := 0
	var restartTimer *time.Timer
	pendingRestart := false
	pendingChange := false

	triggerStart := func() {
		restartCount = 0
		if restartTimer != nil {
			restartTimer.Stop()
			restartTimer = nil
		}
		pendingRestart = false
		startProcess()
	}

	// Start initial process
	triggerStart()

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			killProcess()
			return nil
		case <-sigChan:
			fmt.Println("\nShutting down...")
			killProcess()
			return nil

		case path := <-changeChan:
			fmt.Printf("\nFile change detected: %s\n", path)
			if params.Execute == "" {
				continue
			}

			cmdMutex.Lock()
			isRunning := processRunning
			cmdMutex.Unlock()

			if isRunning {
				if params.PreviousProcess == "kill" {
					killProcess()
					pendingChange = true
				} else {
					// wait mode - mark as pending but let process finish
					pendingChange = true
				}
			} else {
				triggerStart()
			}

		case err := <-processDone:
			if pendingChange {
				pendingChange = false
				triggerStart()
			} else {
				if err != nil {
					// Process exited with error
					// fmt.Printf("Process exited with error: %v\n", err)
				}

				if params.HandleShutdown == "restart" {
					if restartCount < params.MaxRestarts {
						restartCount++
						minBackoff := time.Duration(params.MinBackoffMillis) * time.Millisecond
						maxBackoff := time.Duration(params.MaxBackoffMillis) * time.Millisecond
						backoff := minBackoff * (1 << (restartCount - 1))
						if backoff > maxBackoff {
							backoff = maxBackoff
						}
						fmt.Printf("Restarting in %v (attempt %d/%d)...\n", backoff, restartCount, params.MaxRestarts)

						if restartTimer != nil {
							restartTimer.Stop()
						}
						restartTimer = time.NewTimer(backoff)
						pendingRestart = true
					} else {
						fmt.Println("Max restarts reached. Waiting for file changes.")
					}
				}
			}

		case <-func() <-chan time.Time {
			if restartTimer != nil && pendingRestart {
				return restartTimer.C
			}
			return nil
		}():
			pendingRestart = false
			restartTimer = nil
			startProcess()
		}
	}
}

// globToRegex converts a simple glob pattern to a regex pattern
func globToRegex(glob string) string {
	regex := strings.Builder{}
	regex.WriteString("^")

	for i := 0; i < len(glob); i++ {
		c := glob[i]
		switch c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				// ** matches any number of directories
				regex.WriteString(".*")
				i++ // Skip next *
			} else {
				// * matches anything except /
				regex.WriteString("[^/]*")
			}
		case '?':
			regex.WriteString(".")
		case '.', '+', '(', ')', '|', '^', '$', '@', '%':
			regex.WriteString("\\")
			regex.WriteByte(c)
		case '[', ']':
			regex.WriteByte(c)
		default:
			regex.WriteByte(c)
		}
	}

	regex.WriteString("$")
	return regex.String()
}

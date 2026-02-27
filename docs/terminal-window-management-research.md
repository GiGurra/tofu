# Terminal Window Management Research

Research for implementing Shift+Enter feature in `tclaude session` watch mode:
- First press: Open new terminal window attached to selected session
- Second press: Focus existing window if already open

## Problem

No cross-platform Go library exists for programmatically opening terminal windows or focusing existing windows by name/title.

## Platform-Specific Solutions

### Linux (X11)

**Opening terminals:**
```bash
gnome-terminal -- tmux attach -t session_name
konsole -e tmux attach -t session_name
xterm -e tmux attach -t session_name
alacritty -e tmux attach -t session_name
```

**Setting window title:**
```bash
gnome-terminal --title="tofu-claude-abc123" -- tmux attach -t session_name
```

**Finding/focusing windows:**
```bash
# xdotool - search and activate by window name
xdotool search --name "tofu-claude-abc123" windowactivate

# wmctrl - activate window by name pattern
wmctrl -a "tofu-claude-abc123"

# List all windows
wmctrl -l
xdotool search --name ""
```

**Go libraries for X11:**
- `github.com/jezek/xgb` - Pure Go X11 protocol implementation
- `github.com/BurntSushi/xgbutil` - Higher-level X11 utilities with EWMH support

**Limitations:**
- X11 only - Wayland has no equivalent (security model prevents this)
- Some window managers may ignore focus requests
- Need to detect which terminal emulator user prefers

### macOS

**Opening terminals:**
```bash
# Terminal.app
osascript -e 'tell application "Terminal" to do script "tmux attach -t session_name"'

# iTerm2
osascript -e 'tell application "iTerm2" to create window with default profile command "tmux attach -t session_name"'
```

**Focusing windows:**
```bash
# Focus Terminal.app
osascript -e 'tell application "Terminal" to activate'

# Focus specific window by name (more complex)
osascript -e '
tell application "Terminal"
    activate
    set theWindows to every window whose name contains "tofu-claude-abc123"
    if (count of theWindows) > 0 then
        set frontmost of item 1 of theWindows to true
    end if
end tell
'
```

**Setting window title:**
Terminal titles can be set via escape sequences from within the shell:
```bash
echo -ne "\033]0;tofu-claude-abc123\007"
```

### Windows

**Opening terminals:**
```powershell
# Windows Terminal
wt.exe -w 0 nt --title "tofu-claude-abc123" cmd /c "tmux attach -t session_name"

# cmd.exe
start "tofu-claude-abc123" cmd /c "tmux attach -t session_name"
```

**Focusing windows:**
Requires Win32 API calls (`FindWindow`, `SetForegroundWindow`) or PowerShell:
```powershell
Add-Type @"
using System;
using System.Runtime.InteropServices;
public class Win32 {
    [DllImport("user32.dll")]
    public static extern bool SetForegroundWindow(IntPtr hWnd);
    [DllImport("user32.dll")]
    public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
}
"@
$hwnd = [Win32]::FindWindow($null, "tofu-claude-abc123")
[Win32]::SetForegroundWindow($hwnd)
```

**Go libraries for Windows:**
- `github.com/lxn/win` - Win32 API bindings
- `golang.org/x/sys/windows` - Low-level Windows syscalls

## Proposed Implementation

### Architecture

```
cmd/claude/session/
  terminal/
    terminal.go      # Interface and detection
    terminal_linux.go
    terminal_darwin.go
    terminal_windows.go
```

### Interface

```go
package terminal

// Manager handles opening and focusing terminal windows
type Manager interface {
    // OpenAttached opens a new terminal attached to the tmux session
    // Returns a window identifier for later focus operations
    OpenAttached(tmuxSession, title string) (WindowID, error)

    // FocusWindow brings an existing window to front
    // Returns false if window not found
    FocusWindow(title string) (bool, error)

    // IsWindowOpen checks if a window with the given title exists
    IsWindowOpen(title string) (bool, error)
}

type WindowID string
```

### Detection Strategy

1. Check `$TERM_PROGRAM` env var (set by many terminals)
2. Check for specific terminal processes
3. Fall back to common defaults per platform:
   - Linux: gnome-terminal, konsole, xterm
   - macOS: Terminal.app, iTerm2
   - Windows: Windows Terminal, cmd.exe

### Title Convention

Use predictable titles: `tofu-claude-<session-id>`

This allows:
- Searching for existing windows
- Multiple sessions to have separate windows
- Easy identification by users

### State Tracking

Store opened window info in session state:
```go
type SessionState struct {
    // ... existing fields ...
    TerminalWindowTitle string `json:"terminalWindowTitle,omitempty"`
}
```

## Challenges

1. **Terminal detection**: Many terminals exist, hard to support all
2. **Wayland**: No window focus API (security restriction)
3. **WSL**: Runs Linux but needs Windows terminal management
4. **User preferences**: Users have strong terminal preferences
5. **Focus stealing**: Some systems block focus stealing for security

## Alternative Approaches

### 1. tmux-only (simplest)

Skip external terminal management entirely:
- Shift+Enter just runs `tmux attach` in current terminal
- User manages their own terminal windows
- Works everywhere tmux works

### 2. Configuration-based

Let users configure their terminal command:
```toml
[session]
terminal_command = "gnome-terminal --title={{title}} -- {{command}}"
focus_command = "wmctrl -a {{title}}"
```

### 3. Desktop notifications

Instead of opening windows, send desktop notification with attach command:
```bash
notify-send "Session ready" "Run: tclaude session attach abc123"
```

## Recommendations

1. **Start simple**: Implement tmux-only approach first
2. **Add Linux/X11 support**: Most straightforward with xdotool/wmctrl
3. **Make it configurable**: Let power users customize terminal commands
4. **Document limitations**: Be clear about Wayland, WSL edge cases

## References

- [xdotool manual](https://manpages.ubuntu.com/manpages/trusty/man1/xdotool.1.html)
- [wmctrl](https://linux.die.net/man/1/wmctrl)
- [AppleScript Terminal automation](https://www.samgalope.dev/2024/09/02/how-to-automate-terminal-with-applescript-terminal-automation-on-macos/)
- [xgbutil Go library](https://github.com/BurntSushi/xgbutil)
- [tmux set-titles](https://github.com/tmux/tmux/wiki/Advanced-Use)

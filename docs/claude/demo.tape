# VHS Demo: tofu claude session & conversation management
# Run with: vhs demo.tape

Output demo.gif
Output demo.mp4
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms
Set Shell bash

# Title slide
Type "# tofu claude - Session & Conversation Management"
Enter
Sleep 1s
Type "# Features: tmux sessions, interactive watch modes, search, filtering"
Enter
Sleep 2s
Ctrl+L

# Show conversation list
Type "# 1. List conversations with interactive watch mode"
Enter
Sleep 1s
Type "./tofu claude conv ls -w"
Sleep 500ms
Enter
Sleep 2s

# Search in conversations
Type "/"
Sleep 500ms
Type "plan"
Sleep 1s

# Navigate
Down
Sleep 500ms
Down
Sleep 500ms
Up
Sleep 500ms

# Clear search
Escape
Sleep 1s

# Show help
Type "h"
Sleep 2s
Escape
Sleep 1s

# Quit watch mode
Type "q"
Sleep 1s

# Show global conversation list
Type "# 2. Global conversation search across all projects"
Enter
Sleep 1s
Type "./tofu claude conv ls -g -w"
Sleep 500ms
Enter
Sleep 2s

# Search globally
Type "/"
Sleep 500ms
Type "hello"
Sleep 1.5s
Escape
Sleep 500ms
Escape
Sleep 500ms
Type "q"
Sleep 1s

# Create a session
Type "# 3. Create a new detached session"
Enter
Sleep 1s
Type "./tofu claude session new -d --resume 96ffb05d"
Sleep 500ms
Enter
Sleep 2s

# Show session watch mode
Type "# 4. Session watch mode with status tracking"
Enter
Sleep 1s
Type "./tofu claude session ls -w"
Sleep 500ms
Enter
Sleep 2s

# Search in sessions
Type "/"
Sleep 500ms
Type "tofu"
Sleep 1s
Escape
Sleep 500ms
Escape
Sleep 500ms

# Show help
Type "h"
Sleep 2s
Escape
Sleep 1s

# Show filter menu
Type "f"
Sleep 1.5s
Escape
Sleep 1s

# Quit
Type "q"
Sleep 1s

# Show conversation with session indicator
Type "# 5. Conversation view shows active session indicator"
Enter
Sleep 1s
Type "./tofu claude conv ls -w"
Sleep 500ms
Enter
Sleep 2s

# Point out the indicator
Sleep 2s
Type "q"
Sleep 1s

# Clean up - kill session
Type "# Cleanup: kill the demo session"
Enter
Sleep 500ms
Type "./tofu claude session kill 96ffb05d -f"
Sleep 500ms
Enter
Sleep 1s

# Final message
Type "# Done! See: github.com/GiGurra/tofu"
Enter
Sleep 2s

# Conversation Management 💬

Browse, search, and manage Claude Code conversations.

## Commands

### conv ls

List conversations.

```bash
# List conversations for current project
tofu claude conv ls

# Interactive watch mode
tofu claude conv ls -w

# Global - all projects
tofu claude conv ls -g

# Global interactive
tofu claude conv ls -g -w

# Limit results
tofu claude conv ls -n 10

# Filter by time
tofu claude conv ls --since 7d
tofu claude conv ls --before 2024-01-01
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-w, --watch` | Interactive watch mode |
| `-g, --global` | Search all projects |
| `-n, --count` | Limit number of results |
| `--since` | Show only after this time |
| `--before` | Show only before this time |

### conv search

Search conversation content.

```bash
# Search in current project
tofu claude conv search "authentication"

# Search globally
tofu claude conv search -g "authentication"

# Search with time filter
tofu claude conv search --since 24h "bug fix"

# Full content search (slower, more thorough)
tofu claude conv search --content "specific error message"
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-g, --global` | Search all projects |
| `--content` | Search full conversation content |
| `--since` | Filter by time |
| `--before` | Filter by time |

### conv resume

Resume a conversation in a new Claude session.

```bash
# Resume by ID (creates a new tmux session)
tofu claude conv resume <conv-id>

# Resume detached
tofu claude conv resume -d <conv-id>
```

This is equivalent to `tofu claude session new --resume <conv-id>`.

### conv delete

Delete a conversation.

```bash
# Delete by ID
tofu claude conv delete <conv-id>

# Skip confirmation
tofu claude conv delete -y <conv-id>

# Search globally
tofu claude conv delete -g <conv-id>
```

### conv prune-empty

Delete conversations with no messages.

```bash
# Prune current project
tofu claude conv prune-empty

# Prune globally
tofu claude conv prune-empty -g

# Preview only (dry run)
tofu claude conv prune-empty --dry-run
```

### conv cp / conv mv

Copy or move conversations.

```bash
# Copy a conversation
tofu claude conv cp <conv-id> /path/to/dest

# Move a conversation
tofu claude conv mv <conv-id> /path/to/dest
```

## Interactive Watch Mode

Press `w` or use `-w` flag to enter interactive mode.

### Navigation

| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `PgUp`/`Ctrl+B` | Page up |
| `PgDn`/`Ctrl+F` | Page down |
| `g`/`Home` | Go to first |
| `G`/`End` | Go to last |
| `Enter` | Create/attach to session |
| `q`/`Esc` | Quit |

### Search

| Key | Action |
|-----|--------|
| `/` | Start search |
| `Esc` | Clear search / exit search mode |
| `Ctrl+U` | Clear search input |
| `↑`/`↓` | Exit search and navigate |

Search matches against: title, first prompt, project path, git branch, session ID.

### Actions

| Key | Action |
|-----|--------|
| `Del`/`x` | Delete conversation |
| `r` | Refresh list |
| `h`/`?` | Show help |

### Delete Confirmation

When deleting a conversation:

- **No active session:** `y` to confirm, `n` to cancel
- **Has active session:**
  - `y` - Delete conversation AND stop session
  - `s` - Stop session only (keep conversation)
  - `n` - Cancel

## Session Indicators ⚡

In the conversation list, indicators show session status:

| Indicator | Meaning |
|-----------|---------|
| ⚡ | Conversation has an attached session (someone's watching!) |
| ○ | Conversation has an active session (running in background) |
| (none) | No active session |

## Time Filters

Commands support flexible time formats:

| Format | Example | Description |
|--------|---------|-------------|
| Duration | `24h`, `7d`, `2w` | Relative time |
| Date | `2024-01-15` | Specific date |
| DateTime | `2024-01-15T10:30` | Date and time |

```bash
# Last 24 hours
tofu claude conv ls --since 24h

# Last week
tofu claude conv ls --since 7d

# Before a specific date
tofu claude conv ls --before 2024-01-01

# Date range
tofu claude conv ls --since 2024-01-01 --before 2024-02-01
```

## Session ID Formats

Conversations can be referenced by:

- **Full ID:** `abc12345-def6-7890-abcd-ef1234567890`
- **Short prefix:** `abc12345` (if unique)
- **Autocomplete format:** `abc12345_[project]_prompt...` (from shell completion)

Shell completions automatically provide the full format with context.

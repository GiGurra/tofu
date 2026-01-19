# claude

Claude Code utilities.

## Synopsis

```bash
tofu claude <subcommand> [flags]
```

## Subcommands

- [`conv`](#conv) - Manage Claude Code conversations

---

## conv

Manage Claude Code conversations. Provides utilities to list, search, and manage conversation sessions.

### Synopsis

```bash
tofu claude conv <subcommand> [flags]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List all conversations |
| `search` | Search conversations by text |
| `ai-search` | Search conversations using AI |
| `resume` | Resume a conversation |
| `cp` | Copy a conversation |
| `mv` | Move a conversation |
| `delete` | Delete a conversation |

### Examples

List conversations:

```bash
tofu claude conv list
```

Search conversations:

```bash
tofu claude conv search "authentication bug"
```

AI-powered search:

```bash
tofu claude conv ai-search "how did we fix the login issue"
```

Resume a specific conversation:

```bash
tofu claude conv resume <session-id>
```

Copy a conversation:

```bash
tofu claude conv cp <session-id> <destination>
```

Move a conversation:

```bash
tofu claude conv mv <session-id> <destination>
```

Delete a conversation:

```bash
tofu claude conv delete <session-id>
```

### Time Filters

Commands support time-based filtering:

| Format | Example | Description |
|--------|---------|-------------|
| Duration | `24h`, `7d`, `2w` | Relative time (hours, days, weeks) |
| Date | `2024-01-15` | Specific date |
| DateTime | `2024-01-15T10:30` | Specific date and time |

```bash
# Conversations from the last 7 days
tofu claude conv list --since 7d

# Conversations before a specific date
tofu claude conv list --before 2024-01-01
```

### Session IDs

Sessions can be referenced by:

- Full session ID: `abc123-def456-789...`
- ID prefix: `abc123` (if unique)

### Notes

- Reads from `~/.claude/projects/` directory
- Session IDs support prefix matching
- Works with Claude Code's session index files

# git

Git utilities.

## Synopsis

```bash
tofu git <subcommand> [flags]
```

## Subcommands

- [`sync`](#sync) - Sync git repo(s) to their default branch

---

## sync

Sync git repositories to their default branch. Can sync a single repo or all repos in a directory.

### Synopsis

```bash
tofu git sync [dir] [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prune` | `-p` | Delete all local branches except the default branch | `false` |
| `--dry-run` | `-n` | Show what would be done without making changes | `false` |
| `--pattern` | | Only process directories matching this regex pattern | |
| `--parallel` | `-P` | Number of repos to process in parallel | `10` |
| `--stash` | | Stash uncommitted changes before syncing (pop after) | `false` |
| `--drop` | | Discard uncommitted changes before syncing (DANGEROUS) | `false` |

### Arguments

| Argument | Description |
|----------|-------------|
| `dir` | Parent directory containing git repos (defaults to current) |

### Examples

Sync current directory (if git repo) or all repos in subdirs:

```bash
tofu git sync
```

Sync all repos in a directory:

```bash
tofu git sync ~/projects
```

Sync a single repo:

```bash
tofu git sync ~/projects/my-repo
```

Preview changes without applying:

```bash
tofu git sync --dry-run
```

Sync and delete non-default branches:

```bash
tofu git sync --prune
```

Process repos sequentially:

```bash
tofu git sync --parallel 1
```

Stash changes, sync, then restore:

```bash
tofu git sync --stash
```

Only sync repos matching pattern:

```bash
tofu git sync --pattern "^api-"
```

### Workflow

For each git repository:

1. Checks for uncommitted changes (skips unless `--stash` or `--drop`)
2. Switches to the default branch
3. Pulls the latest changes (`--ff-only`)
4. Optionally prunes non-default local branches

### Output

```
[dry-run] api-service/
  = main already up to date

[dry-run] web-app/
  + Pulled main (3 new commits)
  - Pruned: feature-x, old-branch

==================================================
SYNC REPORT
==================================================

Repositories: 5 total
  Synced:  4
  Skipped: 1

New commits pulled: 12
Branches pruned:    3

CURRENT STATE
--------------------------------------------------

On default branch (4):
  api-service
  web-app
  ...

With uncommitted changes (1):
  my-wip-project
```

### Notes

- Detects default branch automatically (main, master, develop, trunk)
- Skips repos with uncommitted changes by default
- `--stash` and `--drop` are mutually exclusive
- Uses `git pull --ff-only` for safe pulls

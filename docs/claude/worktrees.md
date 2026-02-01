# Git Worktrees

Manage git worktrees for parallel Claude sessions.

## Why Worktrees?

Git worktrees allow you to have multiple branches checked out simultaneously in separate directories. Combined with Claude Code sessions, this enables:

- **Parallel development** - Work on multiple features at once, each with its own Claude session
- **Context isolation** - Each worktree maintains its own conversation history
- **Quick context switching** - Jump between branches without stashing or committing WIP

## Commands

| Command | Description |
|---------|-------------|
| `worktree add <branch>` | Create a new worktree with a Claude session |
| `worktree ls` | List all worktrees |
| `worktree switch <branch>` | Switch to a worktree (requires shell wrapper) |
| `worktree rm <branch>` | Remove a worktree |

## Creating a Worktree

```bash
# Create a new worktree for a feature branch
tofu claude worktree add feat/my-feature

# Create from a specific base branch
tofu claude worktree add feat/my-feature --from-branch develop

# Copy a conversation to the new worktree
tofu claude worktree add feat/my-feature --from-conv abc123

# Create without starting a session
tofu claude worktree add feat/my-feature --detached
```

This will:

1. Create a new branch (if it doesn't exist)
2. Create a worktree at `../<repo>-feat--my-feature`
3. Optionally copy a conversation to the new project
4. Start a Claude session in the new worktree

### Options

| Flag | Description |
|------|-------------|
| `--from-branch` | Base branch to create from (defaults to main/master) |
| `--from-conv` | Conversation ID to copy to the new worktree |
| `--path` | Custom path for the worktree |
| `-d, --detached` | Don't start a Claude session |
| `-g` | Search globally for conversation (with `--from-conv`) |

## Listing Worktrees

```bash
# List all worktrees
tofu claude worktree ls

# Or just
tofu claude worktree
```

Output shows the path, branch, and commit for each worktree:

```
PATH                              BRANCH           COMMIT
/home/user/myrepo                 main             abc1234 (main)
/home/user/myrepo-feat--feature   feat/feature     def5678
```

## Switching Worktrees

The `switch` command outputs a worktree path, which a shell wrapper can use to `cd` to that directory.

```bash
# With shell wrapper installed:
tofu claude worktree switch feat/my-feature

# Aliases also work:
tofu claude worktree s main
tofu claude worktree c develop  # checkout alias
```

### Shell Wrapper Setup

The `switch` command requires a shell wrapper to actually change directories (a subprocess can't change the parent shell's directory). Add one of these to your shell config:

=== "Zsh"

    Add to `~/.zshrc`:
    ```bash
    source /path/to/tofu/scripts/tofu-worktree-switch.zsh
    ```

    Or copy the function directly:
    ```bash
    tofu() {
        if [[ $# -ge 4 && "$1" == "claude" && "$2" == "worktree" && "$3" =~ ^(switch|s|checkout|c)$ ]]; then
            local dir
            dir=$(command tofu "$@" 2>&1)
            local status_code=$?
            if [[ $status_code -eq 0 && -n "$dir" && -d "$dir" ]]; then
                cd "$dir"
            else
                echo "$dir" >&2
                return $status_code
            fi
        else
            command tofu "$@"
        fi
    }
    ```

=== "Bash"

    Add to `~/.bashrc`:
    ```bash
    source /path/to/tofu/scripts/tofu-worktree-switch.bash
    ```

    Or copy the function directly:
    ```bash
    tofu() {
        if [[ $# -ge 4 && "$1" == "claude" && "$2" == "worktree" && "$3" =~ ^(switch|s|checkout|c)$ ]]; then
            local dir
            dir=$(command tofu "$@" 2>&1)
            local status_code=$?
            if [[ $status_code -eq 0 && -n "$dir" && -d "$dir" ]]; then
                cd "$dir"
            else
                echo "$dir" >&2
                return $status_code
            fi
        else
            command tofu "$@"
        fi
    }
    ```

=== "Fish"

    Add to `~/.config/fish/config.fish`:
    ```fish
    source /path/to/tofu/scripts/tofu-worktree-switch.fish
    ```

    Or copy the function directly:
    ```fish
    function tofu
        if test (count $argv) -ge 4; and test "$argv[1]" = "claude"; and test "$argv[2]" = "worktree"; and string match -qr '^(switch|s|checkout|c)$' "$argv[3]"
            set -l dir (command tofu $argv 2>&1)
            set -l status_code $status
            if test $status_code -eq 0; and test -n "$dir"; and test -d "$dir"
                cd "$dir"
            else
                echo "$dir" >&2
                return $status_code
            end
        else
            command tofu $argv
        end
    end
    ```

## Removing Worktrees

```bash
# Remove a worktree by branch name
tofu claude worktree rm feat/my-feature

# Remove by path
tofu claude worktree rm /path/to/worktree

# Force remove (even if dirty)
tofu claude worktree rm feat/my-feature --force

# Also delete the branch
tofu claude worktree rm feat/my-feature --delete-branch
```

### Options

| Flag | Description |
|------|-------------|
| `-f, --force` | Force removal even if worktree has changes |
| `--delete-branch` | Also delete the branch after removing worktree |

## Interactive Mode

The conversation watch mode (`tofu claude conv ls -w`) supports creating worktrees directly:

| Key | Action |
|-----|--------|
| `W` | Create worktree from selected conversation |

This opens a prompt for the branch name and creates a worktree with the selected conversation copied to it.

## Example Workflow

```bash
# Start working on a feature
tofu claude worktree add feat/auth-refactor

# ... work on auth refactor with Claude ...

# Need to fix an urgent bug? Create another worktree
tofu claude worktree add fix/critical-bug --from-branch main

# Switch between them
tofu claude worktree switch feat/auth-refactor
tofu claude worktree switch fix/critical-bug

# Done with the bug fix
tofu claude worktree rm fix/critical-bug --delete-branch

# List what's still active
tofu claude worktree ls
```

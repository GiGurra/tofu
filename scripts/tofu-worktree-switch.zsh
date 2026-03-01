# Zsh shell wrapper for tofu claude worktree switch
#
# This function shadows the tofu binary to intercept 'worktree switch' commands
# and actually change directory to the worktree.
#
# Usage:
#   source /path/to/tofu-worktree-switch.zsh
#
# Then:
#   tofu claude worktree switch feat/my-feature  # cd's to the worktree
#   tofu claude worktree s main                  # short alias
#   tofu claude worktree c main                  # checkout alias

tofu() {
    # Check if this is a worktree switch command
    # Matches: tofu claude worktree (switch|s|checkout|c) <target>
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
        # Pass through to the real tofu command
        command tofu "$@"
    fi
}

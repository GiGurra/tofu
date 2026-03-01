# Fish shell wrapper for tofu claude worktree switch
#
# This function shadows the tofu binary to intercept 'worktree switch' commands
# and actually change directory to the worktree.
#
# Usage:
#   source /path/to/tofu-worktree-switch.fish
#
# Then:
#   tofu claude worktree switch feat/my-feature  # cd's to the worktree
#   tofu claude worktree s main                  # short alias
#   tofu claude worktree c main                  # checkout alias

function tofu
    # Check if this is a worktree switch command
    # Matches: tofu claude worktree (switch|s|checkout|c) <target>
    if test (count $argv) -ge 4
        and test "$argv[1]" = "claude"
        and test "$argv[2]" = "worktree"
        and contains -- "$argv[3]" switch s checkout c

        set -l dir (command tofu $argv 2>&1)
        set -l status_code $status

        if test $status_code -eq 0 -a -n "$dir" -a -d "$dir"
            cd $dir
        else
            # Output error message
            echo $dir >&2
            return $status_code
        end
    else
        # Pass through to the real tofu command
        command tofu $argv
    end
end

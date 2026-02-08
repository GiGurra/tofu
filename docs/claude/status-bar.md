# Status Bar

A rich status bar for Claude Code's statusline feature.

## Overview

Tofu provides a status bar command that Claude Code calls automatically to display contextual information below the input area. It shows model info, workspace details, git links, context usage, subscription limits, and extra usage status.

**Example output:**

```
[Opus 4.6 2.1.37] | /home/user/project | https://github.com/user/project
ctx ████░░░░▒▒ 42% | 5h ░░░░░░░░░░ 8% (3h41m) | 7d ░░░░░░░░░░ 5% (2d9h) | sonnet ░░░░░░░░░░ 0% (4d11h)
extra usage: off
```

## Setup

The easiest way to install:

```bash
tofu claude setup
```

This adds the status bar configuration to `~/.claude/settings.json`. You can also install it manually:

```json
{
  "statusLine": {
    "type": "command",
    "command": "tofu claude status-bar"
  }
}
```

Check if it's installed:

```bash
tofu claude setup --check
```

## What It Shows

### Line 1: Session Info

```
[Model Version] | Dir | Git Links
```

| Element | Description |
|---------|-------------|
| `[Opus 4.6 2.1.37]` | Model name and Claude Code version (cyan) |
| `📂 /path/to/project` | Current working directory |
| `🔗 <url>` | Git repo URL, branch diff URL, and/or PR URL |

**Git links** adapt to context:
- **On default branch:** shows the repo URL
- **On a feature branch:** shows a compare URL (`repo/compare/main...branch`)
- **With an open PR:** appends the PR URL

### Line 2: Usage & Limits

```
ctx <bar> N% | 5h <bar> N% (timer) | 7d <bar> N% (timer) | sonnet <bar> N% (timer)
```

| Element | Description |
|---------|-------------|
| `ctx` | Context window usage with compaction buffer indicator |
| `5h` | 5-hour rate limit utilization and reset timer |
| `7d` | 7-day rate limit utilization and reset timer |
| `sonnet` | 7-day Sonnet limit (premium/max only) |
| `$N.NN` | Session cost (API plan only, hidden on subscription plans) |

**Progress bars** are color-coded:
- Green: normal usage
- Yellow: moderate usage
- Red: high usage

**Context bar** includes a compaction buffer indicator (`▒▒`) showing the ~16.5% reserved for compaction. Color thresholds are adjusted relative to the effective usable space.

**Reset timers** show time until the limit resets: `(45m)`, `(3h30m)`, or `(2d9h)`.

### Line 3: Extra Usage

```
extra usage: off
```

or when enabled:

```
extra usage: on | $12.50 / $100.00 | <bar> 13%
```

Shows the overuse allowance status, credits used vs monthly limit, and utilization.

## Usage Command

You can also check your subscription limits directly:

```bash
# Human-readable output
tofu claude usage

# Raw JSON from the API
tofu claude usage --json
```

## Caching

The status bar caches data to stay fast (it runs after every assistant message):

| Data | Cache Location | TTL |
|------|---------------|-----|
| Git info (repo, branch, PR) | `~/.cache/tofu-claude-git-<hash>.json` | 15 seconds |
| Subscription limits | `~/.cache/tofu-claude-usage.json` | 15 seconds |

- Git cache is **per-repo** (keyed by repo root hash), so parallel sessions in different repos don't interfere
- Usage cache is **shared** since it's account-level data
- All cache writes are **atomic** (write to temp file + rename) to avoid corruption from parallel sessions
- Context window percentage, cost, and model info come fresh from Claude Code on each invocation

## How It Works

Claude Code pipes JSON session data to the status bar command via stdin. The JSON includes model info, version, workspace directory, context window usage, and cost. The status bar combines this with cached git data and subscription limits to render the output.

The command is hidden from `tofu --help` since it's only meant to be called by Claude Code.

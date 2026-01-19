# gh

GitHub utilities.

## Synopsis

```bash
tofu gh <subcommand> [flags]
```

## Subcommands

- [`list-repos`](#list-repos) - List repositories for a user, org, or team
- [`open`](#open) - Open a GitHub repository in the browser

---

## list-repos

List all repositories for a GitHub user or organization, optionally filtered by team.

### Synopsis

```bash
tofu gh list-repos -o <owner> [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--owner` | `-o` | GitHub organization or user name | (required) |
| `--team` | | Team slug/name within the organization | |
| `--visibility` | `-v` | Filter by visibility (can specify multiple) | `all` |
| `--archived` | | Filter by archived status | `all` |
| `--sort` | | Sort repos by field | `full_name` |
| `--direction` | | Sort direction | `asc` |
| `--url` | | Print full GitHub URLs instead of repo names | `false` |

### Visibility Options

- `all` - All repositories
- `public` - Public repositories only
- `private` - Private repositories only
- `internal` - Internal repositories only

### Archived Options

- `all` - All repositories
- `archived` - Only archived repositories
- `not-archived` - Only active repositories

### Sort Options

- `full_name` - Repository name
- `created` - Creation date
- `updated` - Last update date
- `pushed` - Last push date

### Examples

List all repos for an organization:

```bash
tofu gh list-repos -o my-org
```

List repos for a specific team:

```bash
tofu gh list-repos -o my-org --team backend
```

List only public repos:

```bash
tofu gh list-repos -o my-org -v public
```

List with full URLs:

```bash
tofu gh list-repos -o my-org --url
```

List non-archived repos sorted by last push:

```bash
tofu gh list-repos -o my-org --archived not-archived --sort pushed --direction desc
```

### Notes

- Requires `gh` CLI to be installed and authenticated
- Shell completion available for organizations and teams

---

## open

Open a GitHub repository in the default web browser.

### Synopsis

```bash
tofu gh open [dir|url] [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--remote` | `-r` | Git remote name to use | `origin` |

### Arguments

| Argument | Description |
|----------|-------------|
| `dir` | Directory path to a git repository |
| `url` | GitHub URL to open directly |

### Examples

Open current directory's repo:

```bash
tofu gh open
```

Open a specific directory:

```bash
tofu gh open ~/projects/my-repo
```

Open a URL:

```bash
tofu gh open https://github.com/owner/repo
```

Open using a different remote:

```bash
tofu gh open -r upstream
```

Short URL format:

```bash
tofu gh open github.com/owner/repo
```

### Notes

- Auto-detects SSH and HTTPS remote URLs
- Converts SSH URLs to HTTPS for browser
- Works on macOS, Linux, and Windows

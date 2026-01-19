# busy

Look productive with a fake progress display.

## Synopsis

```bash
tofu busy [flags]
```

## Description

Display a fake compilation/build progress. Perfect for looking busy during meetings or while thinking about the actual problem.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--duration` | `-d` | Duration in seconds (0 = infinite) | `0` |
| `--message` | `-m` | Custom status message | (random) |

## Examples

Look busy indefinitely:

```bash
tofu busy
```

Look busy for 60 seconds:

```bash
tofu busy -d 60
```

Custom message:

```bash
tofu busy -m "Optimizing neural networks"
```

## Sample Output

```
⠹ Reticulating splines [████████████░░░░░░░░] 62%
```

## Tasks

The following tasks cycle randomly:

- Compiling
- Linking
- Optimizing
- Resolving dependencies
- Downloading packages
- Building modules
- Running tests
- Generating code
- Analyzing
- Minifying
- Bundling
- Transpiling
- Type checking
- Deploying
- Syncing
- Reticulating splines
- Reversing the polarity
- Calibrating flux capacitor
- Consulting the oracle
- Summoning dependencies

## Notes

- Press Ctrl+C to exit
- Progress bar animates continuously
- Tasks change randomly to simulate real builds

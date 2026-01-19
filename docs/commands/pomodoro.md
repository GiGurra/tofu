# pomodoro

Pomodoro timer for productivity.

## Synopsis

```bash
tofu pomodoro [flags]
```

## Description

A simple pomodoro timer. Work in focused sessions with regular breaks to maintain productivity. Uses the classic Pomodoro Technique: 25 minutes of work followed by a 5-minute break, with a longer break after 4 sessions.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--work` | `-w` | Work duration in minutes | `25` |
| `--break` | `-b` | Break duration in minutes | `5` |
| `--long-break` | `-l` | Long break duration in minutes | `15` |
| `--sessions` | `-n` | Number of sessions before long break | `4` |
| `--continuous` | `-c` | Run continuously (multiple pomodoros) | `false` |

## Examples

Start a standard pomodoro:

```bash
tofu pomodoro
```

Shorter work sessions:

```bash
tofu pomodoro -w 15 -b 3
```

Continuous mode:

```bash
tofu pomodoro -c
```

Custom configuration:

```bash
tofu pomodoro -w 50 -b 10 -l 20 -n 3
```

## Display

```
Pomodoro #1 - Work time! (25 minutes)
Working [████████████████████░░░░░░░░░░] 12:34
```

## Workflow

1. Work session starts (25 min by default)
2. Short break (5 min)
3. Repeat 4 times
4. Long break (15 min)
5. Cycle continues if `-c` flag is set

## Notes

- Terminal bell sounds when sessions end
- Press Ctrl+C to exit
- Progress bar shows time remaining

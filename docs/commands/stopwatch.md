# stopwatch

Simple stopwatch.

## Synopsis

```bash
tofu stopwatch
```

## Description

A terminal stopwatch with lap functionality. Displays hours, minutes, seconds, and milliseconds.

## Controls

| Key | Action |
|-----|--------|
| `Space` | Record a lap |
| `Enter` | Pause/Resume |
| `Q` | Quit |
| `Ctrl+C` | Quit |

## Examples

Start the stopwatch:

```bash
tofu stopwatch
```

## Display

```
STOPWATCH
━━━━━━━━━━━━━━━━━━━━━━━━━━
Space: Lap | Enter: Pause/Resume | Q: Quit

▶  00:01:23.456

Laps:
  #1  00:00:30.123
  #2  00:00:58.456
  #3  00:01:15.789
```

## Features

- Millisecond precision
- Lap recording (Space key)
- Pause/Resume (Enter key)
- Shows last 5 laps
- Status indicator (▶ running, ⏸ paused)

## Notes

- Press Ctrl+C or Q to exit
- Requires terminal with ANSI escape code support
- Uses raw terminal mode for instant key response

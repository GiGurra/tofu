# calendar

Display a calendar.

## Synopsis

```bash
tofu calendar [flags]
```

## Description

Display a terminal calendar with today highlighted. Shows a clean month view.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--month` | `-m` | Month (1-12) | current month |
| `--year` | `-y` | Year | current year |

## Examples

Current month:

```bash
tofu calendar
```

Specific month:

```bash
tofu calendar -m 12
```

Specific month and year:

```bash
tofu calendar -m 7 -y 2024
```

## Sample Output

```
    January 2025
Su Mo Tu We Th Fr Sa
          1  2  3  4
 5  6  7  8  9 10 11
12 13 14 15 16 17 18
19 20 21 22 23 24 25
26 27 28 29 30 31
```

## Features

- Today is highlighted (reverse video)
- Week starts on Sunday
- Clean, minimal output
- Works for any month/year

## Notes

- Today is highlighted only when viewing the current month
- Requires terminal with ANSI escape code support for highlighting

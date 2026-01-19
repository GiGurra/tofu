# clock

Display an analog clock in the terminal.

## Synopsis

```bash
tofu clock [flags]
```

## Description

Shows a beautiful analog clock with hour, minute, and second hands in your terminal. Digital time displayed below.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--size` | `-s` | Clock radius | `12` |

## Examples

Display the clock:

```bash
tofu clock
```

Smaller clock:

```bash
tofu clock -s 8
```

Larger clock:

```bash
tofu clock -s 20
```

## Display

```
           ·  12  ·
       ·            ·
      11              1
    ·                   ·
   ·         ●           ·
  10        ○◎            2
   ·       ∙              ·
    ·                   ·
      9               3
       ·            ·
           ·  6  ·

        14:30:45
```

## Legend

| Symbol | Meaning |
|--------|---------|
| `●` | Hour hand |
| `○` | Minute hand |
| `∙` | Second hand |
| `◎` | Center |
| `·` | Clock face |
| 1-12 | Hour markers |

## Notes

- Updates every 100ms for smooth animation
- Press Ctrl+C to exit
- Requires terminal with ANSI escape code support
- Clock auto-scales based on the size parameter

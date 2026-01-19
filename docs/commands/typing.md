# typing

Typing speed test.

## Synopsis

```bash
tofu typing [flags]
```

## Description

Test your typing speed. Type the displayed words as fast as you can and get detailed statistics including WPM, accuracy, and a rating.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--words` | `-w` | Number of words to type | `25` |

## Examples

Start a typing test:

```bash
tofu typing
```

Shorter test:

```bash
tofu typing -w 10
```

Longer challenge:

```bash
tofu typing -w 50
```

## Gameplay

1. Words are displayed on screen
2. Type them as fast and accurately as you can
3. Press Enter when done
4. Get your results

## Results

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RESULTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Time:      45.2 seconds
Gross WPM: 62
Net WPM:   58
Accuracy:  93.5%
Words:     23/25 correct
```

## Ratings

| Net WPM | Rating |
|---------|--------|
| 80+ | LEGENDARY! Are you a court stenographer? |
| 60-79 | Excellent! You're a typing wizard! |
| 40-59 | Good job! Above average typing speed. |
| 25-39 | Not bad! Keep practicing. |
| < 25 | Keep at it! Practice makes perfect. |

## Word List

Includes common English words plus programming terms:
- Common: the, be, to, and, have, for, not, with, etc.
- Programming: function, variable, class, return, import, async, await, etc.

## Notes

- WPM calculated using standard 5 characters = 1 word
- Net WPM factors in accuracy
- Timer starts when you begin typing

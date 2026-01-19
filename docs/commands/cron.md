# cron

Explain and validate cron expressions.

## Synopsis

```bash
tofu cron <expression> [flags]
```

## Description

Parse cron expressions and show human-readable explanations with upcoming execution times. Supports both 5-field and 6-field (with seconds) cron expressions.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--next` | `-n` | Show next N execution times | `5` |
| `--validate` | `-v` | Only validate, don't explain | `false` |

## Examples

Explain a cron expression:

```bash
tofu cron "0 9 * * 1-5"
```

Explain with seconds:

```bash
tofu cron "0 0 9 * * 1-5"
```

Show next 10 execution times:

```bash
tofu cron -n 10 "*/15 * * * *"
```

Validate only:

```bash
tofu cron -v "0 0 1 * *"
```

## Cron Expression Format

Standard 5-field format:
```
┌───────── minute (0-59)
│ ┌─────── hour (0-23)
│ │ ┌───── day of month (1-31)
│ │ │ ┌─── month (1-12 or jan-dec)
│ │ │ │ ┌─ day of week (0-6, 0=Sunday, or sun-sat)
│ │ │ │ │
* * * * *
```

6-field format (with seconds):
```
┌───────── second (0-59)
│ ┌─────── minute (0-59)
│ │ ┌───── hour (0-23)
│ │ │ ┌─── day of month (1-31)
│ │ │ │ ┌─ month (1-12)
│ │ │ │ │ ┌ day of week (0-6)
│ │ │ │ │ │
* * * * * *
```

## Special Characters

| Character | Meaning |
|-----------|---------|
| `*` | Any value |
| `,` | List separator (e.g., `1,3,5`) |
| `-` | Range (e.g., `1-5`) |
| `/` | Step (e.g., `*/15`) |

## Sample Output

```
Expression: 0 9 * * 1-5

Schedule:
  minute:         0
  hour:           9:00 AM
  day of month:   every day
  month:          every month
  day of week:    Monday through Friday

Next 5 execution times:
  1. Mon, 15 Jan 2024 09:00:00 UTC
  2. Tue, 16 Jan 2024 09:00:00 UTC
  3. Wed, 17 Jan 2024 09:00:00 UTC
  4. Thu, 18 Jan 2024 09:00:00 UTC
  5. Fri, 19 Jan 2024 09:00:00 UTC
```

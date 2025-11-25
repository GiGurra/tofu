# `tofu cron`

Explain and validate cron expressions. Parses standard 5-field and extended 6-field (with seconds) cron expressions, providing human-readable explanations and upcoming execution times.

## Interface

```
> tofu cron --help
Explain and validate cron expressions

Usage:
  tofu cron <expression> [flags]

Flags:
  -n, --next int    Show next N execution times. (default 5)
  -v, --validate    Only validate the expression, don't explain.
  -h, --help        help for cron
```

## Cron Expression Format

### Standard 5-field format

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12 or jan-dec)
│ │ │ │ ┌───────────── day of week (0-6 or sun-sat, 0=Sunday)
│ │ │ │ │
* * * * *
```

### Extended 6-field format (with seconds)

```
┌───────────── second (0-59)
│ ┌───────────── minute (0-59)
│ │ ┌───────────── hour (0-23)
│ │ │ ┌───────────── day of month (1-31)
│ │ │ │ ┌───────────── month (1-12 or jan-dec)
│ │ │ │ │ ┌───────────── day of week (0-6 or sun-sat)
│ │ │ │ │ │
* * * * * *
```

## Special Characters

| Character | Description | Example |
|-----------|-------------|---------|
| `*` | Any value | `* * * * *` (every minute) |
| `,` | List of values | `1,15,30 * * * *` (at minutes 1, 15, 30) |
| `-` | Range of values | `1-5 * * * *` (minutes 1 through 5) |
| `/` | Step values | `*/15 * * * *` (every 15 minutes) |

## Named Values

### Months
`jan`, `feb`, `mar`, `apr`, `may`, `jun`, `jul`, `aug`, `sep`, `oct`, `nov`, `dec`

### Days of Week
`sun`, `mon`, `tue`, `wed`, `thu`, `fri`, `sat`

## Examples

Explain a cron expression:

```
> tofu cron "30 4 * * *"
Expression: 30 4 * * *

Schedule:
  minute:         30
  hour:           4:00 AM
  day of month:   every day-of-month
  month:          every month
  day of week:    every day-of-week

Next 5 execution times:
  1. Mon, 25 Nov 2024 04:30:00 UTC
  2. Tue, 26 Nov 2024 04:30:00 UTC
  3. Wed, 27 Nov 2024 04:30:00 UTC
  4. Thu, 28 Nov 2024 04:30:00 UTC
  5. Fri, 29 Nov 2024 04:30:00 UTC
```

Every 15 minutes:

```
> tofu cron "*/15 * * * *"
Expression: */15 * * * *

Schedule:
  minute:         every 15 minutes starting at 0
  hour:           every hour
  day of month:   every day-of-month
  month:          every month
  day of week:    every day-of-week
...
```

Every Monday at noon:

```
> tofu cron "0 12 * * mon"
Expression: 0 12 * * mon

Schedule:
  minute:         0
  hour:           12:00 PM (noon)
  day of month:   every day-of-month
  month:          every month
  day of week:    Monday
...
```

First day of every month at midnight:

```
> tofu cron "0 0 1 * *"
Expression: 0 0 1 * *

Schedule:
  minute:         0
  hour:           12:00 AM (midnight)
  day of month:   1
  month:          every month
  day of week:    every day-of-week
...
```

Validate only (no explanation):

```
> tofu cron -v "0 0 * * *"
Valid cron expression
```

Show more execution times:

```
> tofu cron -n 10 "0 9 * * 1-5"
Expression: 0 9 * * 1-5
...
Next 10 execution times:
  1. Mon, 25 Nov 2024 09:00:00 UTC
  ...
```

6-field cron with seconds:

```
> tofu cron "30 0 4 * * *"
Expression: 30 0 4 * * *

Schedule:
  second:         30
  minute:         0
  hour:           4:00 AM
  day of month:   every day-of-month
  month:          every month
  day of week:    every day-of-week
...
```

## Common Cron Expressions

| Expression | Description |
|------------|-------------|
| `* * * * *` | Every minute |
| `*/5 * * * *` | Every 5 minutes |
| `0 * * * *` | Every hour |
| `0 0 * * *` | Every day at midnight |
| `0 0 * * 0` | Every Sunday at midnight |
| `0 0 1 * *` | First day of every month |
| `0 0 1 1 *` | Every year on January 1st |
| `0 9-17 * * 1-5` | Every hour 9am-5pm on weekdays |
| `0 0 * * 1-5` | Every weekday at midnight |

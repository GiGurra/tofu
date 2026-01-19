# standup

Periodic reminders to stand up and stretch.

## Synopsis

```bash
tofu standup [flags]
```

## Description

Reminds you to stand up and stretch at regular intervals. Includes random motivational messages and exercise suggestions. Your body will thank you.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--interval` | `-i` | Interval between reminders in minutes | `30` |
| `--quiet` | `-q` | No bell sound | `false` |

## Examples

Start reminders every 30 minutes:

```bash
tofu standup
```

Every 20 minutes:

```bash
tofu standup -i 20
```

Quiet mode (no bell):

```bash
tofu standup -q
```

## Display

```
Standup reminder started! You'll be reminded every 30 minutes.
Press Ctrl+C to stop.

Next reminder in 25:42

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#1: Time to stand up and stretch!
   Try: Roll your shoulders 5 times
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Messages

Random messages include:
- "Time to stand up and stretch!"
- "Your spine called - it wants a break!"
- "Hydration check! Drink some water and stretch!"
- "Your legs are not decorative. Use them!"
- "RSI prevention time! Stretch those wrists!"

## Exercises

Suggested exercises include:
- Stretch your arms above your head
- Roll your shoulders 5 times
- Look at something 20 feet away for 20 seconds
- Do 5 standing squats
- Walk to get a glass of water
- Stretch your fingers and wrists
- Gently stretch your neck side to side

## Notes

- Press Ctrl+C to exit and see total reminder count
- Countdown timer shows time until next reminder
- Terminal bell alerts you (unless `-q` is used)

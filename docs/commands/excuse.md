# excuse

Generate programmer excuses.

## Synopsis

```bash
tofu excuse [flags]
```

## Description

Generate random programmer excuses for when things go wrong. A comprehensive collection of time-tested developer explanations.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--count` | `-n` | Number of excuses to generate | `1` |

## Examples

Get an excuse:

```bash
tofu excuse
# Output: It works on my machine.
```

Get multiple excuses:

```bash
tofu excuse -n 5
# Output:
# That's not a bug, it's a feature.
# It must be a caching issue.
# The tests passed locally.
# It's probably a race condition.
# DNS.
```

## Sample Excuses

- It works on my machine.
- That's not a bug, it's a feature.
- It must be a caching issue.
- Have you tried clearing your cache?
- It worked yesterday.
- Someone must have changed something.
- It's a known issue.
- The tests passed locally.
- It's probably cosmic rays.
- The requirements were unclear.
- We'll fix it in v2.
- Git blame says it wasn't me.
- It's a timezone issue.
- The CI/CD pipeline is flaky.
- Have you tried turning it off and on again?
- The logs don't show anything.
- It only happens under load.
- Someone deployed without telling me.

## Pro Tips

Combine with other commands:

```bash
# Excuse of the day
tofu excuse | tofu cowsay

# Rainbow excuse
tofu excuse | tofu lolcat
```

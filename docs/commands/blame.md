# blame

Randomly blame something for the bug.

## Synopsis

```bash
tofu blame [flags]
```

## Description

When things go wrong, this command will tell you what to blame. Perfect for post-mortems and blameless retrospectives.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--count` | `-n` | Number of things to blame | `1` |

## Examples

Find someone/something to blame:

```bash
tofu blame
# Output: It's clearly the intern.
```

Blame multiple things:

```bash
tofu blame -n 3
# Output:
# It's clearly DNS.
# It's clearly kubernetes.
# It's clearly cosmic rays.
```

## Sample Culprits

- cosmic rays
- the intern
- the previous developer
- mercury retrograde
- kubernetes
- DNS
- the cache
- JavaScript
- NPM
- left-pad
- technical debt
- the sprint deadline
- the meeting that could have been an email
- legacy code
- that one guy who left 3 years ago
- ChatGPT
- tabs vs spaces
- git rebase
- undefined behavior
- off-by-one errors

## Use Cases

- Blameless post-mortems (with humor)
- Rubber duck debugging
- Stress relief during incidents
- Team bonding

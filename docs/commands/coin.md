# coin

Flip a coin.

## Synopsis

```bash
tofu coin [flags]
```

## Description

Flip a virtual coin and get heads or tails. Useful for making decisions or settling arguments.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--count` | `-n` | Number of flips | `1` |
| `--animate` | `-a` | Show flip animation | `false` |

## Examples

Single flip:

```bash
tofu coin
# Output: Heads
```

Multiple flips:

```bash
tofu coin -n 10
# Output:
# Heads
# Tails
# Heads
# Tails
# Tails
# Heads
# ...
```

With animation:

```bash
tofu coin -a
# Shows spinning coin animation before result
```

## Use Cases

- Making binary decisions
- Settling disputes fairly
- Statistical experiments
- When you can't decide what to have for lunch

# pick

Randomly pick from a list.

## Synopsis

```bash
tofu pick [items...] [flags]
```

## Description

Randomly select one or more items from a provided list. Great for making decisions or random selection.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--count` | `-n` | Number of items to pick | `1` |

## Examples

Pick one item:

```bash
tofu pick apple banana orange
# Output: banana
```

Pick multiple items:

```bash
tofu pick -n 2 alice bob charlie dave
# Output:
# charlie
# alice
```

Pick from quoted strings:

```bash
tofu pick "Option A" "Option B" "Option C"
# Output: Option B
```

Pick a random team member for code review:

```bash
tofu pick alice bob charlie dave eve
# Output: charlie
```

Pick lunch destination:

```bash
tofu pick "Burger King" "McDonald's" "Subway" "Taco Bell"
# Output: Subway
```

## Use Cases

- Randomly assigning tasks
- Selecting code reviewers
- Choosing where to eat
- Drawing lottery winners
- Random sampling from a set

# dice

Roll dice using D&D notation.

## Synopsis

```bash
tofu dice [notation] [flags]
```

## Description

Roll dice using standard D&D/tabletop RPG notation. Supports multiple dice, different sides, and modifiers.

## Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `notation` | Dice notation (e.g., `2d6`, `1d20+5`) | `1d6` |

## Examples

Roll a d6:

```bash
tofu dice
# Output: 4
```

Roll a d20:

```bash
tofu dice 1d20
# Output: 17
```

Roll 2d6:

```bash
tofu dice 2d6
# Output: 2d6: [3, 5] = 8
```

Roll with modifier:

```bash
tofu dice 1d20+5
# Output: 1d20+5: [14] + 5 = 19
```

Roll with negative modifier:

```bash
tofu dice 2d6-2
# Output: 2d6-2: [4, 3] - 2 = 5
```

Roll multiple different dice:

```bash
tofu dice 3d8+2
# Output: 3d8+2: [5, 2, 7] + 2 = 16
```

## Notation

Standard D&D dice notation: `NdS+M`

| Component | Description | Example |
|-----------|-------------|---------|
| N | Number of dice | `2` in `2d6` |
| d | Dice separator | always `d` |
| S | Number of sides | `6` in `2d6` |
| +M | Positive modifier (optional) | `+5` in `1d20+5` |
| -M | Negative modifier (optional) | `-2` in `2d6-2` |

## Common Dice

| Notation | Common Use |
|----------|------------|
| `1d4` | Dagger damage |
| `1d6` | Short sword damage |
| `1d8` | Longsword damage |
| `1d10` | Heavy crossbow |
| `1d12` | Greataxe damage |
| `1d20` | Attack rolls, skill checks |
| `1d100` | Percentile rolls |

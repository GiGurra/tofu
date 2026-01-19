# uwu

UwU-ify text.

## Synopsis

```bash
tofu uwu [text] [flags]
```

## Description

Transform text into uwu-speak. Replaces r/l with w, adds "nya", and optionally adds stuttering and emotes.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--stutter` | `-s` | Add stuttering (e.g., h-hello) | `false` |
| `--emotes` | `-e` | Add random emotes | `false` |

## Examples

Basic uwu-ify:

```bash
tofu uwu "Hello World"
# Output: Hewwo Wowwd
```

With stuttering:

```bash
tofu uwu -s "Hello friend"
# Output: H-hewwo fwiend
```

With emotes:

```bash
tofu uwu -e "I love you"
# Output: I wuv you uwu
```

All options:

```bash
tofu uwu -s -e "Hello everyone"
# Output: H-hewwo evewyone (◕ᴗ◕✿)
```

From stdin:

```bash
echo "Error: connection refused" | tofu uwu
# Output: Ewwow: connection wefused
```

UwU-ify logs:

```bash
tail -f /var/log/app.log | tofu uwu
```

## Transformations

| Original | UwU |
|----------|-----|
| r, R | w, W |
| l, L | w, W |
| no, No, NO | nyo, Nyo, NYO |
| na, ne, ni, nu | nya, nye, nyi, nyu |
| ove | uv |

## Emotes

Random emotes include:
- uwu, UwU, owo, OwO
- >w<, ^w^, :3, x3
- (ノ◕ヮ◕)ノ*:・゚✧
- (◕ᴗ◕✿)
- And more!

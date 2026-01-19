# fortune

Tech fortune cookies.

## Synopsis

```bash
tofu fortune [flags]
```

## Description

Display a random tech-related fortune cookie or programming wisdom. Perfect for starting your day or adding to your shell startup.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--category` | `-c` | Category: `tech`, `wisdom`, `debugging`, `all` | `all` |

## Examples

Random fortune:

```bash
tofu fortune
```

Tech jokes only:

```bash
tofu fortune -c tech
```

Wisdom quotes:

```bash
tofu fortune -c wisdom
```

Debugging humor:

```bash
tofu fortune -c debugging
```

## Categories

### Tech (`-c tech`)
- "There are only two hard things in computer science: cache invalidation, naming things, and off-by-one errors."
- "Why do programmers prefer dark mode? Because light attracts bugs."
- "Debugging is like being the detective in a crime movie where you're also the murderer."

### Wisdom (`-c wisdom`)
- "Done is better than perfect."
- "Make it work, make it right, make it fast."
- "Premature optimization is the root of all evil."
- "Code is read more often than it is written."

### Debugging (`-c debugging`)
- "It works on my machine!"
- "It must be a compiler bug."
- "99 little bugs in the code, 99 little bugs. Take one down, patch it around... 127 little bugs in the code."

## Fun Combinations

Add to your shell profile:

```bash
# In .bashrc or .zshrc
tofu fortune
```

Rainbow fortune:

```bash
tofu fortune | tofu lolcat
```

Fortune with cowsay:

```bash
tofu fortune | tofu cowsay -a gopher
```

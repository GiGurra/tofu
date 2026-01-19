# cowsay

Make an ASCII animal say something.

## Synopsis

```bash
tofu cowsay [message] [flags]
```

## Description

Generate an ASCII picture of an animal saying something. A classic Unix tradition, now available cross-platform!

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--animal` | `-a` | Animal: `cow`, `tux`, `tofu`, `gopher`, `cat`, `ghost` | `cow` |
| `--think` | `-t` | Think instead of say (thought bubble) | `false` |

## Examples

Basic cowsay:

```bash
tofu cowsay "Hello, World!"
```

Use a different animal:

```bash
tofu cowsay -a gopher "Go is great!"
```

Make it think:

```bash
tofu cowsay -t "Hmm..."
```

From stdin:

```bash
echo "Piped message" | tofu cowsay
```

## Animals

### cow (default)
```
 ______________
< Hello World! >
 --------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
```

### tux
```
 ______________
< Hello World! >
 --------------
       \
        \
            .--.
           |o_o |
           |:_/ |
          //   \ \
         (|     | )
        /'\_   _/`\
        \___)=(___/
```

### tofu
```
 ______________
< Hello World! >
 --------------
       \
        \  ___________
         | |  TOFU  | |
         | |________| |
          \__________/
           |        |
           |________|
```

### gopher
```
 ______________
< Hello World! >
 --------------
       \
        \
          ʕ◔ϖ◔ʔ
         /    \
        | ^  ^ |
         \    /
          ~~~~
```

### cat
```
 ______________
< Hello World! >
 --------------
       \
        \    /\_/\
         \  ( o.o )
            > ^ <
           /|   |\
          (_|   |_)
```

### ghost
```
 ______________
< Hello World! >
 --------------
       \
        \   ___
          /    \
         | ^  ^ |
         |  __  |
          \    /
           ^^^^
```

## Fun Combinations

```bash
# Fortune cow
tofu fortune | tofu cowsay

# Rainbow cow
tofu cowsay "Moo!" | tofu lolcat

# UwU cow
tofu cowsay "Hello friend" | tofu uwu
```

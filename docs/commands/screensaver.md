# screensaver

Display an animated tofu bowl screensaver.

## Synopsis

```bash
tofu screensaver [flags]
```

## Description

Display an animated tofu bowl with chopsticks in your terminal. Features rising steam, animated chopsticks, and tofu pieces.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--fps` | `-f` | Frames per second (1-60) | `10` |

## Examples

Start the screensaver:

```bash
tofu screensaver
```

Faster animation:

```bash
tofu screensaver -f 20
```

## Sample Animation

```
          ~   ~   ~
         ~ ` ' ~ ` '
        /   \
       /     \
   ___________________
  /  # @ # @ #        \
 |   [##]  @@  [##]    |
 |  ~~~  [##]  ~~~     |
 |    @@   [##]   @@   |
  \                    /
   \__________________/
      |__________|

     ~ T O F U ~
```

## Features

- Rising steam animation
- Chopsticks moving up and down
- Animated tofu pieces
- Auto-centers based on terminal size
- Clean exit on Ctrl+C

## Notes

- Press Ctrl+C to exit
- Requires terminal with ANSI escape code support
- Automatically adapts to terminal size

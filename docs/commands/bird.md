# bird

Play Flappy Tofu - a terminal flappy bird game.

## Synopsis

```bash
tofu bird [flags]
```

## Description

Play Flappy Tofu in your terminal! Guide your tofu through the gaps in the chopsticks. Includes music playback if audio is available.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--fps` | `-f` | Frames per second (5-60) | `15` |
| `--speed` | `-s` | Game speed (1-5) | `2` |

## Controls

| Key | Action |
|-----|--------|
| `SPACE` or `ENTER` | Flap (jump) |
| `q` or `ESC` | Quit |
| `n` | Next song |
| `Ctrl+C` | Exit |

## Examples

Start the game:

```bash
tofu bird
```

Slower speed for beginners:

```bash
tofu bird -s 1
```

Faster frame rate:

```bash
tofu bird -f 30
```

## Gameplay

```
          ^_^
          [#]
          \~/
      ####     ####
      ::::     ::::
      ####     ####
==================
Score: 5 | High Score: 12
```

- Your tofu automatically falls due to gravity
- Press SPACE to flap and gain height
- Navigate through gaps in the chopsticks
- Each gap passed scores a point
- Game speeds up as you progress

## Notes

- Music requires a CGO build with audio support
- On Linux, requires `libasound2-dev` for audio
- Works on any terminal that supports ANSI escape codes

# `tofu bird`

Play Flappy Tofu - a terminal flappy bird game.

## Synopsis

```
tofu bird [flags]
```

## Description

A fun terminal-based Flappy Bird clone. Guide your tofu through the gaps in the chopsticks!

## Controls

- `SPACE` or `ENTER`: Flap (jump)
- `q` or `ESC`: Quit
- `Ctrl+C`: Exit

## Options

- `-f, --fps int`: Frames per second (default 15)
- `-s, --speed int`: Game speed (1-5) (default 2)

## Examples

Start the game with default settings:

```
tofu bird
```

Play at higher speed:

```
tofu bird -s 4
```

Play with smoother animation:

```
tofu bird -f 30
```

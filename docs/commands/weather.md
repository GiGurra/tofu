# weather

Display ASCII weather.

## Synopsis

```bash
tofu weather [location] [flags]
```

## Description

Fetch and display weather using wttr.in. Shows ASCII art weather for any location.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Format: `full`, `short`, `oneline` | `short` |
| `--units` | `-u` | Units: `m` (metric), `u` (imperial) | `m` |

## Examples

Current location weather:

```bash
tofu weather
```

Weather for a city:

```bash
tofu weather "New York"
```

Weather by airport code:

```bash
tofu weather JFK
```

Full detailed forecast:

```bash
tofu weather -f full London
```

One line format:

```bash
tofu weather -f oneline Tokyo
```

Use imperial units:

```bash
tofu weather -u u Chicago
```

## Sample Output

Short format:
```
Weather report: New York

      \   /     Sunny
       .-.      +22(24) °C
    ― (   ) ―   ↑ 15 km/h
       `-'      10 km
      /   \     0.0 mm
```

One line format:
```
New York: ☀️ +22°C
```

## Notes

- Powered by [wttr.in](https://wttr.in)
- If no location specified, uses IP geolocation
- Supports city names, airport codes, and coordinates

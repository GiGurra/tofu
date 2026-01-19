# name

Generate random names.

## Synopsis

```bash
tofu name [flags]
```

## Description

Generate random names in various styles: operation codenames, project names, variable names, or animal-based names.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--style` | `-s` | Name style: `operation`, `project`, `variable`, `animal` | `operation` |
| `--count` | `-n` | Number of names to generate | `1` |

## Examples

Operation codename:

```bash
tofu name
# Output: THUNDERING FALCON
```

Project name:

```bash
tofu name -s project
# Output: crimson-phoenix
```

Variable name:

```bash
tofu name -s variable
# Output: swiftRaven
```

Animal-based name:

```bash
tofu name -s animal
# Output: Energetic Penguin
```

Generate multiple names:

```bash
tofu name -n 5
# Output:
# SILENT EAGLE
# CRIMSON THUNDER
# GOLDEN PHOENIX
# SHADOW WOLF
# ARCTIC STORM
```

## Styles

| Style | Format | Example |
|-------|--------|---------|
| `operation` | ADJECTIVE NOUN (uppercase) | THUNDERING FALCON |
| `project` | adjective-noun (kebab-case) | crimson-phoenix |
| `variable` | adjectiveNoun (camelCase) | swiftRaven |
| `animal` | Adjective Animal (title case) | Energetic Penguin |

## Use Cases

- Naming secret projects
- Generating branch names
- Creating codenames for releases
- Naming test fixtures
- When you're out of creative ideas

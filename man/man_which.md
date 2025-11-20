# `tofu which`

Locate a program in the user's PATH.

## Interface

```
> tofu which --help
Locate a program in the user's PATH

Usage:
  tofu which [programs] [flags]

Flags:
  -h, --help   help for which
```

### Examples

```
> tofu which go
/usr/local/go/bin/go

> tofu which git
/usr/bin/git

> tofu which non-existent-program
non-existent-program not found
```

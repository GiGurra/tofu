# `tofu free`

Display amount of free and used memory in the system.

## Interface

```
> tofu free --help
Display amount of free and used memory in the system

Usage:
  tofu free [flags]

Flags:
  -g, --gigabytes   Display output in gigabytes.
  -m, --megabytes   Display output in megabytes.
  -h, --help        help for free
```

## Description

The `free` command provides a summary of total, used, and free physical and swap memory in the system. By default, values are displayed in kilobytes. You can use flags to show values in megabytes or gigabytes.

## Examples

Display memory usage in kilobytes (default):

```
> tofu free
            total       used       free     shared buff/cache   available
Mem:     16777216    8388608    4194304    1048576    1572864    6291456 KiB
Swap:     4194304    2097152    2097152 KiB
```

Display memory usage in megabytes:

```
> tofu free -m
            total       used       free     shared buff/cache   available
Mem:        16384       8192       4096       1024       1536       6144 MiB
Swap:        4096       2048       2048 MiB
```

Display memory usage in gigabytes:

```
> tofu free -g
            total       used       free     shared buff/cache   available
Mem:           16          8          4          1          2          6
Swap:           4          2          2
```

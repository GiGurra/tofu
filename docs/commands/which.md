# which

Locate a program in the user's PATH.

## Synopsis

```bash
tofu which <programs...>
```

## Description

Locate the executable file associated with the given program name by searching the PATH environment variable.

## Examples

Find a single program:

```bash
tofu which go
```

Find multiple programs:

```bash
tofu which python node java
```

## Sample Output

```
$ tofu which go
/usr/local/go/bin/go

$ tofu which python node
/usr/bin/python
/usr/local/bin/node

$ tofu which nonexistent
nonexistent not found
```

## Notes

- Returns the full path to the executable
- Exits with code 1 if any program is not found
- On Windows, also searches the current directory for executables with common extensions (.exe, .bat, .cmd, etc.)

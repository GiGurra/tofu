# `tofu uuid`

Generates Universally Unique Identifiers (UUIDs).

## Interface

```
> tofu uuid --help
Generate UUIDs

Usage:
  tofu uuid [flags]

Flags:
  -n, --count int         Number of UUIDs to generate. (default 1)
  -v, --version int       UUID Version (1, 3, 4, 5, 6, 7). (default 4)
  -s, --namespace string  Namespace for v3/v5 (dns, url, oid, x500, or UUID string).
  -d, --name string       Data/Name for v3/v5 generation.
  -h, --help              help for uuid
```

## Versions

- **v1**: Time-based (includes MAC address).
- **v3**: MD5 hash-based (deterministic). Requires namespace and name.
- **v4**: Random (default).
- **v5**: SHA-1 hash-based (deterministic). Requires namespace and name.
- **v6**: Reordered time-based (sortable).
- **v7**: Unix Epoch time-based (sortable, modern standard).

### Examples

Generate a random UUID (v4):

```
> tofu uuid
a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11
```

Generate 5 time-ordered UUIDs (v7):

```
> tofu uuid -v 7 -n 5
018b8a5c-9a3b-70a1-8c1d-123456789abc
...
```

Generate a deterministic UUID (v5) based on DNS namespace:

```
> tofu uuid -v 5 -s dns -d "example.com"
cf4cc793-16f9-5206-b61c-326936016076
```

Generate a deterministic UUID (v3) with custom namespace:

```
> tofu uuid -v 3 -s a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11 -d "my-data"
```

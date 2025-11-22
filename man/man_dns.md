# `tofu dns`

Lookup DNS records for a hostname.

## Interface

```
> tofu dns --help
Lookup DNS records

Usage:
  tofu dns [hostname] [flags]

Flags:
  -s, --server string   DNS server to use (e.g. 8.8.8.8). Defaults to 8.8.8.8:53 (default "8.8.8.8:53")
  -t, --types string    Comma-separated list of record types to query (A,AAAA,CNAME,MX,TXT,NS,PTR). Default: all common types
  -o, --use-os          Use OS resolver instead of direct query (ignores --server) (default false)
  -h, --help            help for dns
```

### Examples

**Default Lookup (Google DNS)**

```
> tofu dns google.com
Server:  8.8.8.8:53
Address: google.com

A Records:
  142.250.74.14

AAAA Records:
  2607:f8b0:400a:80b::200e

MX Records:
  10 smtp.google.com.

TXT Records:
  v=spf1 include:_spf.google.com ~all
```

**Use Specific Server**

```
> tofu dns github.com --server 1.1.1.1
Server:  1.1.1.1:53
Address: github.com

...
```

**Use OS Resolver**

```
> tofu dns localhost --use-os
Server:  OS Default
Address: localhost

A Records:
  127.0.0.1
```

**Specific Record Types**

```
> tofu dns gmail.com --types MX,TXT
Server:  8.8.8.8:53
Address: gmail.com

MX Records:
  5 gmail-smtp-in.l.google.com.
  ...

TXT Records:
  v=spf1 redirect=_spf.google.com
```

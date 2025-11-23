# `tofu jwt`

Decode and inspect JSON Web Tokens (JWT).

## Interface

```
> tofu jwt --help
Decode and inspect JWT tokens

Usage:
  tofu jwt [token] [flags]

Flags:
  -h, --help   help for jwt
```

## Description

The `jwt` command decodes a JSON Web Token (JWT) into its three components: Header, Payload (Claims), and Signature. It pretty-prints the Header and Payload as JSON.

The token can be provided as a command-line argument or via standard input (pipe).

## Examples

Decode a token passed as an argument:

```
> tofu jwt eyJhbGci...
Token:
eyJhbGci...

Header:
{
  "alg": "HS256",
  "typ": "JWT"
}

Payload:
{
  "sub": "1234567890",
  "name": "John Doe",
  "iat": 1516239022
}

Signature (raw):
...
```

Decode a token from stdin:

```
> echo "eyJhbGci..." | tofu jwt
...
```

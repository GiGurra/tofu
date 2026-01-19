# jwt

Work with JSON Web Tokens (JWT).

## Synopsis

```bash
tofu jwt [token]              # Decode (default)
tofu jwt decode [token]       # Decode and inspect
tofu jwt create [flags]       # Create a new token
tofu jwt validate [token]     # Validate a token
```

## Description

Decode, create, and validate JSON Web Tokens. Supports HMAC (HS256/384/512), RSA (RS256/384/512), and ECDSA (ES256/384/512) algorithms.

## Commands

### decode

Decode and inspect a JWT token.

```bash
tofu jwt decode <token>
tofu jwt <token>              # decode is default
echo "eyJ..." | tofu jwt      # from stdin
```

### create

Create a new signed JWT token.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--algorithm` | `-a` | Signing algorithm | `HS256` |
| `--secret` | `-s` | Secret key or path to key file | |
| `--subject` | | Subject claim (sub) | |
| `--issuer` | | Issuer claim (iss) | |
| `--audience` | | Audience claim (aud) | |
| `--expires-in` | `-e` | Expiration time (e.g., 1h, 24h, 7d) | |
| `--no-exp` | | Create without expiration | `false` |
| `--not-before` | | Not before time | |
| `--issued-at` | | Include iat claim | `true` |
| `--id` | | JWT ID (jti) | |
| `--claims` | `-c` | Additional claims as JSON | |

### validate

Validate a JWT token.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--secret` | `-s` | Secret key or path to public key | |
| `--issuer` | | Expected issuer | |
| `--audience` | | Expected audience | |
| `--subject` | | Expected subject | |

## Examples

Decode a token:

```bash
tofu jwt eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Create a simple token:

```bash
tofu jwt create -s "my-secret" --subject "user123" -e 24h
```

Create with custom claims:

```bash
tofu jwt create -s "my-secret" -e 1h -c '{"role":"admin"}'
```

Validate a token:

```bash
tofu jwt validate -s "my-secret" eyJhbGci...
```

Validate with expected issuer:

```bash
tofu jwt validate -s "my-secret" --issuer "myapp" eyJhbGci...
```

## Sample Output

Decode output:
```
Token:
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Header:
{
  "alg": "HS256",
  "typ": "JWT"
}

Payload:
{
  "sub": "user123",
  "iat": 1704067200,
  "exp": 1704153600
}

Time Claims:
  exp: 2024-01-02T00:00:00Z (valid for 23h59m)
  iat: 2024-01-01T00:00:00Z (issued 1h ago)

Signature (raw):
SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

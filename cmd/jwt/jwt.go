package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
)

type Params struct {
}

type DecodeParams struct {
	Token string `pos:"true" optional:"true" help:"JWT token to decode."`
}

type CreateParams struct {
	Algorithm string `short:"a" help:"Signing algorithm (HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512, none)." default:"HS256"`
	Secret    string `short:"s" help:"Secret key for HMAC algorithms or path to private key file for RSA/ECDSA." optional:"true"`
	Subject   string `help:"Subject claim (sub)." optional:"true"`
	Issuer    string `help:"Issuer claim (iss)." optional:"true"`
	Audience  string `help:"Audience claim (aud). Comma-separated for multiple values." optional:"true"`
	ExpiresIn string `short:"e" help:"Token expiration time (e.g., 1h, 24h, 7d, 30m). Required unless --no-exp is set." optional:"true"`
	NoExp     bool   `help:"Create token without expiration."`
	NotBefore string `help:"Not before time (e.g., 1h, 30m) from now, or RFC3339 timestamp." optional:"true"`
	IssuedAt  bool   `help:"Include issued at claim (iat)." default:"true"`
	ID        string `help:"JWT ID claim (jti)." optional:"true"`
	Claims    string `short:"c" help:"Additional claims as JSON object (e.g., '{\"role\":\"admin\"}')." optional:"true"`
}

type ValidateParams struct {
	Token    string `pos:"true" optional:"true" help:"JWT token to validate."`
	Secret   string `short:"s" help:"Secret key for HMAC algorithms or path to public key file for RSA/ECDSA." optional:"true"`
	Issuer   string `help:"Expected issuer (iss) claim." optional:"true"`
	Audience string `help:"Expected audience (aud) claim." optional:"true"`
	Subject  string `help:"Expected subject (sub) claim." optional:"true"`
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jwt",
		Short: "Work with JWT tokens (decode, create, validate)",
		Long: `Work with JSON Web Tokens (JWT).

Subcommands:
  decode    Decode and inspect a JWT token (default if no subcommand)
  create    Create a new signed JWT token
  validate  Validate a JWT token's signature and claims`,
	}

	cmd.AddCommand(decodeCmd())
	cmd.AddCommand(createCmd())
	cmd.AddCommand(validateCmd())

	// Make decode the default action when no subcommand is provided
	cmd.Run = func(cmd *cobra.Command, args []string) {
		token := ""
		if len(args) > 0 {
			token = args[0]
		} else {
			// Read from stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
					os.Exit(1)
				}
				token = strings.TrimSpace(string(data))
			}
		}
		if token == "" {
			_ = cmd.Help()
			return
		}
		// Run decode by default
		if err := runJwtDecode(token); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	return cmd
}

func decodeCmd() *cobra.Command {
	return boa.CmdT[DecodeParams]{
		Use:   "decode [token]",
		Short: "Decode and inspect a JWT token",
		Long: `Decode and inspect a JSON Web Token (JWT).
The token can be provided as an argument or via standard input.
Displays the decoded Header, Payload (Claims), and the Signature.`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *DecodeParams, cmd *cobra.Command, args []string) {
			token := params.Token
			if token == "" || token == "-" {
				// Read from stdin
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					data, err := io.ReadAll(os.Stdin)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
						os.Exit(1)
					}
					token = strings.TrimSpace(string(data))
				}
			}
			if token == "" {
				_ = cmd.Help()
				os.Exit(1)
			}
			if err := runJwtDecode(token); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func createCmd() *cobra.Command {
	return boa.CmdT[CreateParams]{
		Use:   "create",
		Short: "Create a new signed JWT token",
		Long: `Create a new signed JSON Web Token (JWT).

Examples:
  # Create a simple token with HMAC-SHA256
  tofu jwt create -s "my-secret" --subject "user123" -e 24h

  # Create a token with custom claims
  tofu jwt create -s "my-secret" -e 1h -c '{"role":"admin","permissions":["read","write"]}'

  # Create a token with RSA signing
  tofu jwt create -a RS256 -s /path/to/private.pem --issuer "myapp" -e 7d

  # Create an unsigned token (not recommended for production)
  tofu jwt create -a none --subject "test" -e 1h`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *CreateParams, cmd *cobra.Command, args []string) {
			if err := runJwtCreate(params, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func validateCmd() *cobra.Command {
	return boa.CmdT[ValidateParams]{
		Use:   "validate [token]",
		Short: "Validate a JWT token",
		Long: `Validate a JSON Web Token (JWT).

Checks:
  - Signature validity (if secret/key provided)
  - Expiration time (exp)
  - Not before time (nbf)
  - Issued at time (iat)
  - Issuer claim (iss) if --issuer specified
  - Audience claim (aud) if --audience specified
  - Subject claim (sub) if --subject specified

Examples:
  # Validate signature and expiration
  tofu jwt validate -s "my-secret" eyJhbGci...

  # Validate with expected issuer
  tofu jwt validate -s "my-secret" --issuer "myapp" eyJhbGci...

  # Validate from stdin
  echo "eyJhbGci..." | tofu jwt validate -s "my-secret"`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *ValidateParams, cmd *cobra.Command, args []string) {
			token := params.Token
			if token == "" || token == "-" {
				// Read from stdin
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					data, err := io.ReadAll(os.Stdin)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
						os.Exit(1)
					}
					token = strings.TrimSpace(string(data))
				}
			}
			if token == "" {
				_ = cmd.Help()
				os.Exit(1)
			}
			if err := runJwtValidate(params, token, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runJwtDecode(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts (Header.Payload.Signature), found %d", len(parts))
	}

	fmt.Println("Token:")
	fmt.Println(token)
	fmt.Println()

	// Header
	header, err := decodeSegment(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode header: %w", err)
	}
	fmt.Println("Header:")
	printJSON(header)
	fmt.Println()

	// Payload
	payload, err := decodeSegment(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}
	fmt.Println("Payload:")
	printJSON(payload)

	// Check for time-based claims and display human-readable info
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err == nil {
		printTimeClaims(claims)
	}
	fmt.Println()

	// Signature
	fmt.Println("Signature (raw):")
	fmt.Println(parts[2])

	return nil
}

func printTimeClaims(claims map[string]interface{}) {
	now := time.Now()
	fmt.Println()
	fmt.Println("Time Claims:")

	if exp, ok := getNumericClaim(claims, "exp"); ok {
		expTime := time.Unix(exp, 0)
		if expTime.Before(now) {
			fmt.Printf("  exp: %s (EXPIRED %s ago)\n", expTime.Format(time.RFC3339), now.Sub(expTime).Round(time.Second))
		} else {
			fmt.Printf("  exp: %s (valid for %s)\n", expTime.Format(time.RFC3339), expTime.Sub(now).Round(time.Second))
		}
	}

	if nbf, ok := getNumericClaim(claims, "nbf"); ok {
		nbfTime := time.Unix(nbf, 0)
		if nbfTime.After(now) {
			fmt.Printf("  nbf: %s (not yet valid, wait %s)\n", nbfTime.Format(time.RFC3339), nbfTime.Sub(now).Round(time.Second))
		} else {
			fmt.Printf("  nbf: %s (active since %s ago)\n", nbfTime.Format(time.RFC3339), now.Sub(nbfTime).Round(time.Second))
		}
	}

	if iat, ok := getNumericClaim(claims, "iat"); ok {
		iatTime := time.Unix(iat, 0)
		fmt.Printf("  iat: %s (issued %s ago)\n", iatTime.Format(time.RFC3339), now.Sub(iatTime).Round(time.Second))
	}
}

func getNumericClaim(claims map[string]interface{}, key string) (int64, bool) {
	val, ok := claims[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func runJwtCreate(params *CreateParams, stdout io.Writer) error {
	// Validate algorithm
	method := getSigningMethod(params.Algorithm)
	if method == nil {
		return fmt.Errorf("unsupported algorithm: %s", params.Algorithm)
	}

	// Check if secret is required
	if params.Algorithm != "none" && params.Secret == "" {
		return fmt.Errorf("secret (-s) is required for algorithm %s", params.Algorithm)
	}

	// Validate expiration requirement
	if !params.NoExp && params.ExpiresIn == "" {
		return fmt.Errorf("expiration (-e) is required. Use --no-exp to create a token without expiration")
	}

	// Build claims
	claims := jwt.MapClaims{}

	if params.Subject != "" {
		claims["sub"] = params.Subject
	}
	if params.Issuer != "" {
		claims["iss"] = params.Issuer
	}
	if params.Audience != "" {
		audiences := strings.Split(params.Audience, ",")
		if len(audiences) == 1 {
			claims["aud"] = audiences[0]
		} else {
			claims["aud"] = audiences
		}
	}
	if params.ID != "" {
		claims["jti"] = params.ID
	}

	now := time.Now()

	if params.IssuedAt {
		claims["iat"] = now.Unix()
	}

	if params.ExpiresIn != "" {
		exp, err := parseDuration(params.ExpiresIn)
		if err != nil {
			return fmt.Errorf("invalid expiration time: %w", err)
		}
		claims["exp"] = now.Add(exp).Unix()
	}

	if params.NotBefore != "" {
		nbf, err := parseTimeOrDuration(params.NotBefore, now)
		if err != nil {
			return fmt.Errorf("invalid not-before time: %w", err)
		}
		claims["nbf"] = nbf.Unix()
	}

	// Parse additional claims
	if params.Claims != "" {
		var additionalClaims map[string]interface{}
		if err := json.Unmarshal([]byte(params.Claims), &additionalClaims); err != nil {
			return fmt.Errorf("invalid claims JSON: %w", err)
		}
		for k, v := range additionalClaims {
			claims[k] = v
		}
	}

	// Create token
	token := jwt.NewWithClaims(method, claims)

	// Get signing key
	key, err := getSigningKey(params.Algorithm, params.Secret)
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}

	// Sign token
	tokenString, err := token.SignedString(key)
	if err != nil {
		return fmt.Errorf("failed to sign token: %w", err)
	}

	fmt.Fprintln(stdout, tokenString)
	return nil
}

func runJwtValidate(params *ValidateParams, tokenString string, stdout io.Writer) error {
	// Build parser options
	var parserOpts []jwt.ParserOption

	if params.Issuer != "" {
		parserOpts = append(parserOpts, jwt.WithIssuer(params.Issuer))
	}
	if params.Audience != "" {
		parserOpts = append(parserOpts, jwt.WithAudience(params.Audience))
	}
	if params.Subject != "" {
		parserOpts = append(parserOpts, jwt.WithSubject(params.Subject))
	}

	// Parse token without verification first to get the algorithm
	parser := jwt.NewParser(parserOpts...)

	var token *jwt.Token
	var err error

	if params.Secret == "" {
		// Parse without signature verification
		token, _, err = parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			return fmt.Errorf("failed to parse token: %w", err)
		}
		fmt.Fprintln(stdout, "⚠️  Warning: No secret provided, signature not verified")
		fmt.Fprintln(stdout)
	} else {
		// Parse with signature verification
		token, err = parser.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			alg := t.Method.Alg()
			return getVerifyingKey(alg, params.Secret)
		})
		if err != nil {
			return formatValidationError(err)
		}
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("failed to extract claims")
	}

	// Display validation results
	fmt.Fprintln(stdout, "Validation Results:")
	fmt.Fprintln(stdout, "-------------------")

	// Signature
	if params.Secret != "" {
		fmt.Fprintln(stdout, "✓ Signature: valid")
	}

	// Algorithm
	fmt.Fprintf(stdout, "✓ Algorithm: %s\n", token.Method.Alg())

	// Time-based validations
	now := time.Now()

	if exp, ok := getNumericClaim(claims, "exp"); ok {
		expTime := time.Unix(exp, 0)
		if expTime.Before(now) {
			fmt.Fprintf(stdout, "✗ Expiration: EXPIRED at %s (%s ago)\n", expTime.Format(time.RFC3339), now.Sub(expTime).Round(time.Second))
		} else {
			fmt.Fprintf(stdout, "✓ Expiration: valid until %s (%s remaining)\n", expTime.Format(time.RFC3339), expTime.Sub(now).Round(time.Second))
		}
	} else {
		fmt.Fprintln(stdout, "- Expiration: not set")
	}

	if nbf, ok := getNumericClaim(claims, "nbf"); ok {
		nbfTime := time.Unix(nbf, 0)
		if nbfTime.After(now) {
			fmt.Fprintf(stdout, "✗ Not Before: not yet valid until %s (wait %s)\n", nbfTime.Format(time.RFC3339), nbfTime.Sub(now).Round(time.Second))
		} else {
			fmt.Fprintf(stdout, "✓ Not Before: valid since %s\n", nbfTime.Format(time.RFC3339))
		}
	} else {
		fmt.Fprintln(stdout, "- Not Before: not set")
	}

	if iat, ok := getNumericClaim(claims, "iat"); ok {
		iatTime := time.Unix(iat, 0)
		fmt.Fprintf(stdout, "✓ Issued At: %s (%s ago)\n", iatTime.Format(time.RFC3339), now.Sub(iatTime).Round(time.Second))
	} else {
		fmt.Fprintln(stdout, "- Issued At: not set")
	}

	// Standard claims
	if iss, ok := claims["iss"].(string); ok {
		if params.Issuer != "" && iss == params.Issuer {
			fmt.Fprintf(stdout, "✓ Issuer: %s (matches expected)\n", iss)
		} else if params.Issuer != "" {
			fmt.Fprintf(stdout, "✗ Issuer: %s (expected: %s)\n", iss, params.Issuer)
		} else {
			fmt.Fprintf(stdout, "✓ Issuer: %s\n", iss)
		}
	}

	if sub, ok := claims["sub"].(string); ok {
		if params.Subject != "" && sub == params.Subject {
			fmt.Fprintf(stdout, "✓ Subject: %s (matches expected)\n", sub)
		} else if params.Subject != "" {
			fmt.Fprintf(stdout, "✗ Subject: %s (expected: %s)\n", sub, params.Subject)
		} else {
			fmt.Fprintf(stdout, "✓ Subject: %s\n", sub)
		}
	}

	if aud, ok := claims["aud"]; ok {
		fmt.Fprintf(stdout, "✓ Audience: %v\n", aud)
	}

	if jti, ok := claims["jti"].(string); ok {
		fmt.Fprintf(stdout, "✓ JWT ID: %s\n", jti)
	}

	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Token is valid ✓")

	return nil
}

func formatValidationError(err error) error {
	switch {
	case strings.Contains(err.Error(), "token is expired"):
		return fmt.Errorf("token has expired (exp claim)")
	case strings.Contains(err.Error(), "token is not valid yet"):
		return fmt.Errorf("token is not yet valid (nbf claim)")
	case strings.Contains(err.Error(), "signature is invalid"):
		return fmt.Errorf("invalid signature: the token signature does not match")
	case strings.Contains(err.Error(), "token has invalid claims"):
		return fmt.Errorf("token has invalid claims: %w", err)
	default:
		return err
	}
}

func getSigningMethod(alg string) jwt.SigningMethod {
	switch strings.ToUpper(alg) {
	case "HS256":
		return jwt.SigningMethodHS256
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	case "RS256":
		return jwt.SigningMethodRS256
	case "RS384":
		return jwt.SigningMethodRS384
	case "RS512":
		return jwt.SigningMethodRS512
	case "ES256":
		return jwt.SigningMethodES256
	case "ES384":
		return jwt.SigningMethodES384
	case "ES512":
		return jwt.SigningMethodES512
	case "NONE":
		return jwt.SigningMethodNone
	default:
		return nil
	}
}

func getSigningKey(alg string, secret string) (interface{}, error) {
	switch strings.ToUpper(alg) {
	case "HS256", "HS384", "HS512":
		return []byte(secret), nil
	case "RS256", "RS384", "RS512":
		// Try to read as file first
		keyData, err := os.ReadFile(secret)
		if err != nil {
			// If not a file, treat as raw PEM data
			keyData = []byte(secret)
		}
		return jwt.ParseRSAPrivateKeyFromPEM(keyData)
	case "ES256", "ES384", "ES512":
		keyData, err := os.ReadFile(secret)
		if err != nil {
			keyData = []byte(secret)
		}
		return jwt.ParseECPrivateKeyFromPEM(keyData)
	case "NONE":
		return jwt.UnsafeAllowNoneSignatureType, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", alg)
	}
}

func getVerifyingKey(alg string, secret string) (interface{}, error) {
	switch strings.ToUpper(alg) {
	case "HS256", "HS384", "HS512":
		return []byte(secret), nil
	case "RS256", "RS384", "RS512":
		keyData, err := os.ReadFile(secret)
		if err != nil {
			keyData = []byte(secret)
		}
		return jwt.ParseRSAPublicKeyFromPEM(keyData)
	case "ES256", "ES384", "ES512":
		keyData, err := os.ReadFile(secret)
		if err != nil {
			keyData = []byte(secret)
		}
		return jwt.ParseECPublicKeyFromPEM(keyData)
	case "NONE":
		return jwt.UnsafeAllowNoneSignatureType, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", alg)
	}
}

func parseDuration(s string) (time.Duration, error) {
	// Check for day suffix
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func parseTimeOrDuration(s string, now time.Time) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try duration
	d, err := parseDuration(s)
	if err != nil {
		return time.Time{}, err
	}
	return now.Add(d), nil
}

func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}

func printJSON(data []byte) {
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", "  "); err != nil {
		// Fallback if not valid JSON, just print string
		fmt.Println(string(data))
	} else {
		fmt.Println(out.String())
	}
}

// runJwt is kept for backward compatibility with tests
func runJwt(token string) error {
	return runJwtDecode(token)
}

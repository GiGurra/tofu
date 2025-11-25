package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJwtDecode(t *testing.T) {
	// Create a dummy JWT
	header := `{"alg":"HS256","typ":"JWT"}`
	payload := `{"sub":"1234567890","name":"John Doe","iat":1516239022}`
	signature := "dummy_signature"

	encHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	encPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	token := fmt.Sprintf("%s.%s.%s", encHeader, encPayload, signature)

	// Test valid token
	if err := runJwtDecode(token); err != nil {
		t.Errorf("runJwtDecode failed on valid token: %v", err)
	}

	// Test invalid format (too few parts)
	if err := runJwtDecode("part1.part2"); err == nil {
		t.Error("runJwtDecode should fail on invalid token format (too few parts)")
	}

	// Test invalid base64 in header
	if err := runJwtDecode("invalid_base64$.payload.sig"); err == nil {
		t.Error("runJwtDecode should fail on invalid base64 in header")
	}
}

// Keep backward compatibility test
func TestJwtCommand(t *testing.T) {
	header := `{"alg":"HS256","typ":"JWT"}`
	payload := `{"sub":"1234567890","name":"John Doe","iat":1516239022}`
	signature := "dummy_signature"

	encHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	encPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	token := fmt.Sprintf("%s.%s.%s", encHeader, encPayload, signature)

	// Test backward compatibility function
	if err := runJwt(token); err != nil {
		t.Errorf("runJwt failed on valid token: %v", err)
	}
}

func TestJwtCreate(t *testing.T) {
	tests := []struct {
		name    string
		params  *JwtCreateParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "simple HS256 token",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				Subject:   "user123",
				ExpiresIn: "1h",
				IssuedAt:  true,
			},
			wantErr: false,
		},
		{
			name: "token with all standard claims",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				Subject:   "user123",
				Issuer:    "test-issuer",
				Audience:  "test-audience",
				ExpiresIn: "24h",
				ID:        "unique-id-123",
				IssuedAt:  true,
			},
			wantErr: false,
		},
		{
			name: "token with custom claims",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				ExpiresIn: "1h",
				Claims:    `{"role":"admin","permissions":["read","write"]}`,
			},
			wantErr: false,
		},
		{
			name: "token with multiple audiences",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				ExpiresIn: "1h",
				Audience:  "aud1,aud2,aud3",
			},
			wantErr: false,
		},
		{
			name: "token with day duration",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				ExpiresIn: "7d",
			},
			wantErr: false,
		},
		{
			name: "missing secret for HS256",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				ExpiresIn: "1h",
			},
			wantErr: true,
			errMsg:  "secret (-s) is required",
		},
		{
			name: "missing expiration",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
			},
			wantErr: true,
			errMsg:  "expiration (-e) is required",
		},
		{
			name: "no expiration with flag",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				NoExp:     true,
			},
			wantErr: false,
		},
		{
			name: "invalid algorithm",
			params: &JwtCreateParams{
				Algorithm: "INVALID",
				Secret:    "test-secret",
				ExpiresIn: "1h",
			},
			wantErr: true,
			errMsg:  "unsupported algorithm",
		},
		{
			name: "invalid claims JSON",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				ExpiresIn: "1h",
				Claims:    "invalid json",
			},
			wantErr: true,
			errMsg:  "invalid claims JSON",
		},
		{
			name: "invalid expiration duration",
			params: &JwtCreateParams{
				Algorithm: "HS256",
				Secret:    "test-secret",
				ExpiresIn: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid expiration time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runJwtCreate(tt.params, &buf)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the token is valid JWT format
			tokenStr := strings.TrimSpace(buf.String())
			parts := strings.Split(tokenStr, ".")
			if len(parts) != 3 {
				t.Errorf("expected 3 parts in JWT, got %d", len(parts))
			}
		})
	}
}

func TestJwtValidate(t *testing.T) {
	secret := "test-secret-key"

	// Create a valid token for testing
	createValidToken := func(claims jwt.MapClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte(secret))
		return tokenStr
	}

	now := time.Now()

	tests := []struct {
		name    string
		token   string
		params  *JwtValidateParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid token",
			token: createValidToken(jwt.MapClaims{
				"sub": "user123",
				"exp": now.Add(time.Hour).Unix(),
				"iat": now.Unix(),
			}),
			params: &JwtValidateParams{
				Secret: secret,
			},
			wantErr: false,
		},
		{
			name: "valid token with issuer check",
			token: createValidToken(jwt.MapClaims{
				"sub": "user123",
				"iss": "test-issuer",
				"exp": now.Add(time.Hour).Unix(),
			}),
			params: &JwtValidateParams{
				Secret: secret,
				Issuer: "test-issuer",
			},
			wantErr: false,
		},
		{
			name: "expired token",
			token: createValidToken(jwt.MapClaims{
				"sub": "user123",
				"exp": now.Add(-time.Hour).Unix(),
			}),
			params: &JwtValidateParams{
				Secret: secret,
			},
			wantErr: true,
			errMsg:  "expired",
		},
		{
			name: "not yet valid token (nbf)",
			token: createValidToken(jwt.MapClaims{
				"sub": "user123",
				"exp": now.Add(2 * time.Hour).Unix(),
				"nbf": now.Add(time.Hour).Unix(),
			}),
			params: &JwtValidateParams{
				Secret: secret,
			},
			wantErr: true,
			errMsg:  "not yet valid",
		},
		{
			name: "invalid signature",
			token: createValidToken(jwt.MapClaims{
				"sub": "user123",
				"exp": now.Add(time.Hour).Unix(),
			}),
			params: &JwtValidateParams{
				Secret: "wrong-secret",
			},
			wantErr: true,
			errMsg:  "signature",
		},
		{
			name: "wrong issuer",
			token: createValidToken(jwt.MapClaims{
				"sub": "user123",
				"iss": "actual-issuer",
				"exp": now.Add(time.Hour).Unix(),
			}),
			params: &JwtValidateParams{
				Secret: secret,
				Issuer: "expected-issuer",
			},
			wantErr: true,
		},
		{
			name: "validation without secret (warning only)",
			token: createValidToken(jwt.MapClaims{
				"sub": "user123",
				"exp": now.Add(time.Hour).Unix(),
			}),
			params:  &JwtValidateParams{},
			wantErr: false, // Should succeed but with warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runJwtValidate(tt.params, tt.token, &buf)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check output contains expected validation results
			output := buf.String()
			if !strings.Contains(output, "Validation Results:") {
				t.Error("expected output to contain 'Validation Results:'")
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"1h", time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"24h", 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"30d", 30 * 24 * time.Hour, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetSigningMethod(t *testing.T) {
	tests := []struct {
		alg      string
		expected jwt.SigningMethod
	}{
		{"HS256", jwt.SigningMethodHS256},
		{"hs256", jwt.SigningMethodHS256}, // case insensitive
		{"HS384", jwt.SigningMethodHS384},
		{"HS512", jwt.SigningMethodHS512},
		{"RS256", jwt.SigningMethodRS256},
		{"ES256", jwt.SigningMethodES256},
		{"none", jwt.SigningMethodNone},
		{"NONE", jwt.SigningMethodNone},
		{"INVALID", nil},
	}

	for _, tt := range tests {
		t.Run(tt.alg, func(t *testing.T) {
			result := getSigningMethod(tt.alg)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateAndValidateRoundTrip(t *testing.T) {
	secret := "my-test-secret"

	// Create a token
	createParams := &JwtCreateParams{
		Algorithm: "HS256",
		Secret:    secret,
		Subject:   "user123",
		Issuer:    "test-app",
		ExpiresIn: "1h",
		IssuedAt:  true,
		Claims:    `{"role":"admin"}`,
	}

	var createBuf bytes.Buffer
	err := runJwtCreate(createParams, &createBuf)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	tokenStr := strings.TrimSpace(createBuf.String())

	// Validate the token
	validateParams := &JwtValidateParams{
		Secret:  secret,
		Subject: "user123",
		Issuer:  "test-app",
	}

	var validateBuf bytes.Buffer
	err = runJwtValidate(validateParams, tokenStr, &validateBuf)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	output := validateBuf.String()
	if !strings.Contains(output, "Token is valid") {
		t.Error("expected 'Token is valid' in output")
	}
	if !strings.Contains(output, "Signature: valid") {
		t.Error("expected signature to be valid")
	}
}

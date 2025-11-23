package cmd

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestJwtCommand(t *testing.T) {
	// Create a dummy JWT
	header := `{"alg":"HS256","typ":"JWT"}`
	payload := `{"sub":"1234567890","name":"John Doe","iat":1516239022}`
	signature := "dummy_signature"

	encHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	encPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	token := fmt.Sprintf("%s.%s.%s", encHeader, encPayload, signature)

	// Test valid token
	if err := runJwt(token); err != nil {
		t.Errorf("runJwt failed on valid token: %v", err)
	}

	// Test invalid format (too few parts)
	if err := runJwt("part1.part2"); err == nil {
		t.Error("runJwt should fail on invalid token format (too few parts)")
	}

	// Test invalid base64 in header
	if err := runJwt("invalid_base64$.payload.sig"); err == nil {
		t.Error("runJwt should fail on invalid base64 in header")
	}
}

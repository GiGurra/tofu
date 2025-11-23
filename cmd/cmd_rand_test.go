package cmd

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
)

func TestGenerateRandom(t *testing.T) {
	// Test Int
	params := &RandParams{
		Type: "int",
		Min:  10,
		Max:  20,
	}
	val, err := generateRandom(params)
	if err != nil {
		t.Errorf("int generation failed: %v", err)
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("int generation produced non-int: %s", val)
	}
	if n < 10 || n > 20 {
		t.Errorf("int generation out of range: %d", n)
	}

	// Test Str
	params = &RandParams{
		Type:    "str",
		Length:  10,
		Charset: "a",
	}
	val, err = generateRandom(params)
	if err != nil {
		t.Errorf("str generation failed: %v", err)
	}
	if len(val) != 10 {
		t.Errorf("str generation wrong length: %d", len(val))
	}
	if val != "aaaaaaaaaa" {
		t.Errorf("str generation wrong content: %s", val)
	}

	// Test Hex
	params = &RandParams{
		Type:   "hex",
		Length: 4, // 4 bytes -> 8 hex chars
	}
	val, err = generateRandom(params)
	if err != nil {
		t.Errorf("hex generation failed: %v", err)
	}
	if len(val) != 8 {
		t.Errorf("hex generation wrong length: %d", len(val))
	}
	if _, err := hex.DecodeString(val); err != nil {
		t.Errorf("hex generation produced invalid hex: %s", val)
	}

	// Test Base64
	params = &RandParams{
		Type:   "base64",
		Length: 3, // 3 bytes -> 4 base64 chars
	}
	val, err = generateRandom(params)
	if err != nil {
		t.Errorf("base64 generation failed: %v", err)
	}
	if len(val) != 4 {
		t.Errorf("base64 generation wrong length: %d", len(val))
	}
	if _, err := base64.StdEncoding.DecodeString(val); err != nil {
		t.Errorf("base64 generation produced invalid base64: %s", val)
	}

	// Test Password
	params = &RandParams{
		Type:   "password",
		Length: 12,
	}
	val, err = generateRandom(params)
	if err != nil {
		t.Errorf("password generation failed: %v", err)
	}
	if len(val) != 12 {
		t.Errorf("password generation wrong length: %d (expected 12)", len(val))
	}

	// Test Phrase
	params = &RandParams{
		Type:       "phrase",
		Length:     3,
		Separator:  "-",
		Capitalize: "all",
	}
	val, err = generateRandom(params)
	if err != nil {
		t.Errorf("phrase generation failed: %v", err)
	}
	// "Word-Word-Word"
	// Count separators
	if strings.Count(val, "-") != 2 {
		t.Errorf("phrase separator count wrong: %s", val)
	}
}

func TestRandCmd(t *testing.T) {
	// Just verify it runs without error
	params := &RandParams{
		Type:   "str",
		Length: 5,
		Count:  2,
	}
	if err := runRand(params); err != nil {
		t.Errorf("runRand failed: %v", err)
	}
}
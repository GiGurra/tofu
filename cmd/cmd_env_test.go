package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestListEnv(t *testing.T) {
	// Set a known variable for testing
	os.Setenv("TOFU_TEST_VAR", "test_value")
	defer os.Unsetenv("TOFU_TEST_VAR")

	params := &EnvParams{
		Format: "plain",
		Sort:   true,
	}

	// Just verify it runs without error
	err := listEnv(params)
	if err != nil {
		t.Errorf("listEnv failed: %v", err)
	}
}

func TestListEnvWithFilter(t *testing.T) {
	os.Setenv("TOFU_TEST_AAA", "value1")
	os.Setenv("TOFU_TEST_BBB", "value2")
	defer os.Unsetenv("TOFU_TEST_AAA")
	defer os.Unsetenv("TOFU_TEST_BBB")

	params := &EnvParams{
		Format: "plain",
		Filter: "TOFU_TEST",
		Sort:   true,
	}

	err := listEnv(params)
	if err != nil {
		t.Errorf("listEnv with filter failed: %v", err)
	}
}

func TestRunEnvGet(t *testing.T) {
	os.Setenv("TOFU_GET_TEST", "hello_world")
	defer os.Unsetenv("TOFU_GET_TEST")

	params := &EnvParams{
		Get: "TOFU_GET_TEST",
	}

	// Capture stdout would require more setup, just test no error
	// The actual output goes to stdout
	err := runEnv(params)
	if err != nil {
		t.Errorf("runEnv --get failed: %v", err)
	}
}

func TestRunEnvGetNotSet(t *testing.T) {
	params := &EnvParams{
		Get: "TOFU_NONEXISTENT_VAR_12345",
	}

	err := runEnv(params)
	if err == nil {
		t.Error("expected error for non-existent variable")
	}
	if !strings.Contains(err.Error(), "not set") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunEnvSet(t *testing.T) {
	params := &EnvParams{
		Set: "TOFU_SET_TEST=new_value",
	}

	err := runEnv(params)
	if err != nil {
		t.Errorf("runEnv --set failed: %v", err)
	}

	value := os.Getenv("TOFU_SET_TEST")
	if value != "new_value" {
		t.Errorf("expected 'new_value', got '%s'", value)
	}

	os.Unsetenv("TOFU_SET_TEST")
}

func TestRunEnvSetInvalidFormat(t *testing.T) {
	params := &EnvParams{
		Set: "INVALID_NO_EQUALS",
	}

	err := runEnv(params)
	if err == nil {
		t.Error("expected error for invalid --set format")
	}
	if !strings.Contains(err.Error(), "KEY=VALUE") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunEnvUnset(t *testing.T) {
	os.Setenv("TOFU_UNSET_TEST", "to_be_removed")

	params := &EnvParams{
		Unset: "TOFU_UNSET_TEST",
	}

	err := runEnv(params)
	if err != nil {
		t.Errorf("runEnv --unset failed: %v", err)
	}

	_, exists := os.LookupEnv("TOFU_UNSET_TEST")
	if exists {
		t.Error("variable should have been unset")
	}
}

func TestOutputFormats(t *testing.T) {
	os.Setenv("TOFU_FORMAT_TEST", "format_value")
	defer os.Unsetenv("TOFU_FORMAT_TEST")

	formats := []string{"plain", "json", "shell", "powershell"}

	for _, format := range formats {
		params := &EnvParams{
			Format: format,
			Filter: "TOFU_FORMAT_TEST",
			Sort:   true,
		}

		err := listEnv(params)
		if err != nil {
			t.Errorf("listEnv with format %s failed: %v", format, err)
		}
	}
}

func TestKeysOnly(t *testing.T) {
	os.Setenv("TOFU_KEYS_TEST", "some_value")
	defer os.Unsetenv("TOFU_KEYS_TEST")

	params := &EnvParams{
		Format: "plain",
		Filter: "TOFU_KEYS_TEST",
		Keys:   true,
		Sort:   true,
	}

	err := listEnv(params)
	if err != nil {
		t.Errorf("listEnv with keys only failed: %v", err)
	}
}

func TestValuesOnly(t *testing.T) {
	os.Setenv("TOFU_VALUES_TEST", "some_value")
	defer os.Unsetenv("TOFU_VALUES_TEST")

	params := &EnvParams{
		Format: "plain",
		Filter: "TOFU_VALUES_TEST",
		Values: true,
		Sort:   true,
	}

	err := listEnv(params)
	if err != nil {
		t.Errorf("listEnv with values only failed: %v", err)
	}
}

func TestNoEmpty(t *testing.T) {
	os.Setenv("TOFU_EMPTY_TEST", "")
	defer os.Unsetenv("TOFU_EMPTY_TEST")

	params := &EnvParams{
		Format:  "plain",
		Filter:  "TOFU_EMPTY_TEST",
		NoEmpty: true,
		Sort:    true,
	}

	err := listEnv(params)
	if err != nil {
		t.Errorf("listEnv with no-empty failed: %v", err)
	}
}

func TestEnvCmd(t *testing.T) {
	// Verify the command is created without error
	cmd := EnvCmd()
	if cmd == nil {
		t.Error("EnvCmd returned nil")
	}
	if cmd.Name() != "env" {
		t.Errorf("expected Name()='env', got '%s'", cmd.Name())
	}
}

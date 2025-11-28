package time

import (
	"testing"
	stdtime "time"
)

func TestParseTime(t *testing.T) {
	// 1. Unix Timestamp (Seconds)
	// 1698393600 -> 2023-10-27 08:00:00 UTC
	ts := "1698393600"
	parsed, err := parseTime(ts)
	if err != nil {
		t.Fatalf("Failed to parse unix seconds: %v", err)
	}
	if parsed.Unix() != 1698393600 {
		t.Errorf("Expected 1698393600, got %d", parsed.Unix())
	}

	// 2. Unix Timestamp (Millis)
	// 1698393600000
	tsMillis := "1698393600000"
	parsed, err = parseTime(tsMillis)
	if err != nil {
		t.Fatalf("Failed to parse unix millis: %v", err)
	}
	if parsed.Unix() != 1698393600 {
		t.Errorf("Expected 1698393600 from millis, got %d", parsed.Unix())
	}

	// 3. RFC3339
	rfc := "2023-10-27T10:00:00+02:00"
	parsed, err = parseTime(rfc)
	if err != nil {
		t.Fatalf("Failed to parse RFC3339: %v", err)
	}
	// Check UTC stdtime
	utc := parsed.UTC()
	if utc.Year() != 2023 || utc.Month() != stdtime.October || utc.Day() != 27 || utc.Hour() != 8 {
		t.Errorf("Incorrect time parsed from RFC3339: %v", parsed)
	}

	// 4. DateOnly
	date := "2023-10-27"
	parsed, err = parseTime(date)
	if err != nil {
		t.Fatalf("Failed to parse DateOnly: %v", err)
	}
	if parsed.Year() != 2023 || parsed.Month() != stdtime.October || parsed.Day() != 27 {
		t.Errorf("Incorrect date parsed: %v", parsed)
	}

	// 5. Invalid
	_, err = parseTime("invalid-time")
	if err == nil {
		t.Errorf("Should fail for invalid time")
	}
}

package cron

import (
	"testing"
	"time"
)

func TestParseCronExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"every minute", "* * * * *", false},
		{"specific time", "30 4 * * *", false},
		{"with day of week", "0 0 * * 0", false},
		{"with month", "0 0 1 1 *", false},
		{"range", "0-30 * * * *", false},
		{"step", "*/5 * * * *", false},
		{"list", "0,15,30,45 * * * *", false},
		{"complex", "0 0 1,15 * 1-5", false},
		{"6 fields with seconds", "0 30 4 * * *", false},
		{"month names", "0 0 1 jan *", false},
		{"day names", "0 0 * * mon", false},
		{"too few fields", "* * *", true},
		{"too many fields", "* * * * * * *", true},
		{"invalid value", "60 * * * *", true},
		{"invalid range", "5-2 * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCronExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCronExpression(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
			}
		})
	}
}

func TestParseField(t *testing.T) {
	tests := []struct {
		name       string
		field      string
		fieldName  string
		min, max   int
		wantValues []int
		wantErr    bool
	}{
		{"wildcard", "*", "minute", 0, 59, nil, false},
		{"single value", "5", "minute", 0, 59, []int{5}, false},
		{"range", "1-5", "minute", 0, 59, []int{1, 2, 3, 4, 5}, false},
		{"step", "*/15", "minute", 0, 59, []int{0, 15, 30, 45}, false},
		{"list", "1,3,5", "minute", 0, 59, []int{1, 3, 5}, false},
		{"range with step", "0-30/10", "minute", 0, 59, []int{0, 10, 20, 30}, false},
		{"out of range", "60", "minute", 0, 59, nil, true},
		{"invalid step", "*/0", "minute", 0, 59, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf, err := parseField(tt.field, tt.fieldName, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if tt.wantValues == nil {
				if cf.Values != nil {
					t.Errorf("parseField() got values %v, want nil (wildcard)", cf.Values)
				}
			} else {
				if len(cf.Values) != len(tt.wantValues) {
					t.Errorf("parseField() got %v values, want %v", cf.Values, tt.wantValues)
					return
				}
				for i, v := range cf.Values {
					if v != tt.wantValues[i] {
						t.Errorf("parseField() values[%d] = %d, want %d", i, v, tt.wantValues[i])
					}
				}
			}
		})
	}
}

func TestMatchesCron(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		time  time.Time
		match bool
	}{
		{
			"every minute matches",
			"* * * * *",
			time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			true,
		},
		{
			"specific minute matches",
			"30 * * * *",
			time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			true,
		},
		{
			"specific minute no match",
			"30 * * * *",
			time.Date(2024, 1, 15, 10, 15, 0, 0, time.UTC),
			false,
		},
		{
			"specific hour matches",
			"0 10 * * *",
			time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			true,
		},
		{
			"day of week matches (Monday)",
			"0 0 * * 1",
			time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // Monday
			true,
		},
		{
			"day of week no match",
			"0 0 * * 1",
			time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), // Sunday
			false,
		},
		{
			"month matches",
			"0 0 1 1 *",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			true,
		},
		{
			"month no match",
			"0 0 1 1 *",
			time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCronExpression(tt.expr)
			if err != nil {
				t.Fatalf("failed to parse expression: %v", err)
			}
			got := matchesCron(expr, tt.time)
			if got != tt.match {
				t.Errorf("matchesCron() = %v, want %v", got, tt.match)
			}
		})
	}
}

func TestGetNextExecutions(t *testing.T) {
	expr, err := parseCronExpression("0 * * * *") // Every hour at minute 0
	if err != nil {
		t.Fatalf("failed to parse expression: %v", err)
	}

	from := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	results := getNextExecutions(expr, from, 3)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	expected := []time.Time{
		time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC),
	}

	for i, exp := range expected {
		if !results[i].Equal(exp) {
			t.Errorf("result[%d] = %v, want %v", i, results[i], exp)
		}
	}
}

func TestExplainField(t *testing.T) {
	tests := []struct {
		name        string
		field       *CronField
		displayName string
		contains    string
	}{
		{
			"wildcard",
			&CronField{Name: "minute", Values: nil},
			"minute",
			"every minute",
		},
		{
			"single value",
			&CronField{Name: "minute", Values: []int{30}},
			"minute",
			"30",
		},
		{
			"day name",
			&CronField{Name: "day-of-week", Values: []int{1}},
			"day of week",
			"Monday",
		},
		{
			"month name",
			&CronField{Name: "month", Values: []int{1}},
			"month",
			"January",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := explainField(tt.field, tt.displayName)
			if !containsString(result, tt.contains) {
				t.Errorf("explainField() = %q, want it to contain %q", result, tt.contains)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCronCmd(t *testing.T) {
	cmd := Cmd()
	if cmd == nil {
		t.Error("CronCmd returned nil")
	}
	if cmd.Name() != "cron" {
		t.Errorf("expected Name()='cron', got '%s'", cmd.Name())
	}
}

func TestMonthAndDayNames(t *testing.T) {
	// Test month names
	expr, err := parseCronExpression("0 0 1 jan *")
	if err != nil {
		t.Fatalf("failed to parse with month name: %v", err)
	}
	if len(expr.Month.Values) != 1 || expr.Month.Values[0] != 1 {
		t.Errorf("month 'jan' should parse to 1, got %v", expr.Month.Values)
	}

	// Test day names
	expr, err = parseCronExpression("0 0 * * mon")
	if err != nil {
		t.Fatalf("failed to parse with day name: %v", err)
	}
	if len(expr.DayOfWeek.Values) != 1 || expr.DayOfWeek.Values[0] != 1 {
		t.Errorf("day 'mon' should parse to 1, got %v", expr.DayOfWeek.Values)
	}
}

func TestSixFieldCron(t *testing.T) {
	expr, err := parseCronExpression("30 0 4 * * *")
	if err != nil {
		t.Fatalf("failed to parse 6-field cron: %v", err)
	}

	if !expr.HasSeconds {
		t.Error("expected HasSeconds to be true")
	}

	if expr.Second == nil || len(expr.Second.Values) != 1 || expr.Second.Values[0] != 30 {
		t.Errorf("expected second=30, got %v", expr.Second)
	}

	if expr.Minute == nil || len(expr.Minute.Values) != 1 || expr.Minute.Values[0] != 0 {
		t.Errorf("expected minute=0, got %v", expr.Minute)
	}

	if expr.Hour == nil || len(expr.Hour.Values) != 1 || expr.Hour.Values[0] != 4 {
		t.Errorf("expected hour=4, got %v", expr.Hour)
	}
}

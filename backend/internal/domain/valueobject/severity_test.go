package valueobject

import "testing"

func TestNewSeverity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Severity
		wantErr bool
	}{
		{"low", "low", SeverityLow, false},
		{"medium", "medium", SeverityMedium, false},
		{"high", "high", SeverityHigh, false},
		{"critical", "critical", SeverityCritical, false},
		{"invalid", "extreme", Severity(""), true},
		{"empty string", "", Severity(""), true},
		{"uppercase", "Low", Severity(""), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewSeverity(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if s != tc.want {
					t.Errorf("NewSeverity(%q) = %q, want %q", tc.input, s, tc.want)
				}
			}
		})
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		s    Severity
		want string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
		{SeverityCritical, "critical"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if tc.s.String() != tc.want {
				t.Errorf("String() = %q, want %q", tc.s.String(), tc.want)
			}
		})
	}
}

func TestSeverity_Level(t *testing.T) {
	tests := []struct {
		name  string
		s     Severity
		level int
	}{
		{"low level is 1", SeverityLow, 1},
		{"medium level is 2", SeverityMedium, 2},
		{"high level is 3", SeverityHigh, 3},
		{"critical level is 4", SeverityCritical, 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.s.Level() != tc.level {
				t.Errorf("Level() = %d, want %d for %q", tc.s.Level(), tc.level, tc.s)
			}
		})
	}
}

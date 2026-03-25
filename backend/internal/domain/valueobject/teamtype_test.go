package valueobject

import "testing"

func TestNewTeamType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TeamType
		wantErr bool
	}{
		{"stream-aligned", "stream-aligned", StreamAligned, false},
		{"platform", "platform", Platform, false},
		{"enabling", "enabling", Enabling, false},
		{"complicated-subsystem", "complicated-subsystem", ComplicatedSubsystem, false},
		{"invalid value", "invalid", TeamType(""), true},
		{"empty string", "", TeamType(""), true},
		{"uppercase stream-aligned", "Stream-Aligned", TeamType(""), true},
		{"uppercase platform", "PLATFORM", TeamType(""), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt, err := NewTeamType(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if tt != tc.want {
					t.Errorf("NewTeamType(%q) = %q, want %q", tc.input, tt, tc.want)
				}
			}
		})
	}
}

func TestTeamType_String(t *testing.T) {
	tests := []struct {
		tt   TeamType
		want string
	}{
		{StreamAligned, "stream-aligned"},
		{Platform, "platform"},
		{Enabling, "enabling"},
		{ComplicatedSubsystem, "complicated-subsystem"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if tc.tt.String() != tc.want {
				t.Errorf("String() = %q, want %q", tc.tt.String(), tc.want)
			}
		})
	}
}

func TestTeamType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		tt    TeamType
		valid bool
	}{
		{"stream-aligned is valid", StreamAligned, true},
		{"platform is valid", Platform, true},
		{"enabling is valid", Enabling, true},
		{"complicated-subsystem is valid", ComplicatedSubsystem, true},
		{"empty is invalid", TeamType(""), false},
		{"unknown is invalid", TeamType("unknown"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.tt.IsValid() != tc.valid {
				t.Errorf("IsValid() = %v, want %v for %q", tc.tt.IsValid(), tc.valid, tc.tt)
			}
		})
	}
}

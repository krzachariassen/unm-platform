package valueobject

import "testing"

func TestNewMappingStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MappingStatus
		wantErr bool
	}{
		{"asserted", "asserted", Asserted, false},
		{"inferred", "inferred", Inferred, false},
		{"candidate", "candidate", Candidate, false},
		{"deprecated", "deprecated", Deprecated, false},
		{"invalid", "unknown", MappingStatus(""), true},
		{"empty string", "", MappingStatus(""), true},
		{"uppercase", "Asserted", MappingStatus(""), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms, err := NewMappingStatus(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if ms != tc.want {
					t.Errorf("NewMappingStatus(%q) = %q, want %q", tc.input, ms, tc.want)
				}
			}
		})
	}
}

func TestMappingStatus_String(t *testing.T) {
	tests := []struct {
		ms   MappingStatus
		want string
	}{
		{Asserted, "asserted"},
		{Inferred, "inferred"},
		{Candidate, "candidate"},
		{Deprecated, "deprecated"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if tc.ms.String() != tc.want {
				t.Errorf("String() = %q, want %q", tc.ms.String(), tc.want)
			}
		})
	}
}

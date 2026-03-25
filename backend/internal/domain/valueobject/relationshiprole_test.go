package valueobject

import "testing"

func TestNewRelationshipRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RelationshipRole
		wantErr bool
	}{
		{"primary", "primary", Primary, false},
		{"supporting", "supporting", Supporting, false},
		{"consuming", "consuming", Consuming, false},
		{"empty string (no role)", "", RelationshipRole(""), false},
		{"invalid", "observer", RelationshipRole(""), true},
		{"uppercase", "Primary", RelationshipRole(""), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr, err := NewRelationshipRole(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if rr != tc.want {
					t.Errorf("NewRelationshipRole(%q) = %q, want %q", tc.input, rr, tc.want)
				}
			}
		})
	}
}

func TestRelationshipRole_String(t *testing.T) {
	tests := []struct {
		rr   RelationshipRole
		want string
	}{
		{Primary, "primary"},
		{Supporting, "supporting"},
		{Consuming, "consuming"},
		{RelationshipRole(""), ""},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if tc.rr.String() != tc.want {
				t.Errorf("String() = %q, want %q", tc.rr.String(), tc.want)
			}
		})
	}
}

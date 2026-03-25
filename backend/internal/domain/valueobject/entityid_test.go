package valueobject

import "testing"

func TestNewEntityID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid id", "actor-1", false},
		{"valid id with spaces inside", "my entity", false},
		{"empty string", "", true},
		{"whitespace only spaces", "   ", true},
		{"whitespace only tabs", "\t\t", true},
		{"whitespace mixed", " \t ", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := NewEntityID(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if id.String() != tc.input {
					t.Errorf("String() = %q, want %q", id.String(), tc.input)
				}
			}
		})
	}
}

func TestEntityID_IsZero(t *testing.T) {
	var zero EntityID
	if !zero.IsZero() {
		t.Error("zero value EntityID should return true for IsZero()")
	}

	id, err := NewEntityID("some-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.IsZero() {
		t.Error("non-zero EntityID should return false for IsZero()")
	}
}

func TestEntityID_Equality(t *testing.T) {
	id1, _ := NewEntityID("actor-1")
	id2, _ := NewEntityID("actor-1")
	id3, _ := NewEntityID("actor-2")

	if id1 != id2 {
		t.Error("EntityIDs with same value should be equal")
	}
	if id1 == id3 {
		t.Error("EntityIDs with different values should not be equal")
	}
}

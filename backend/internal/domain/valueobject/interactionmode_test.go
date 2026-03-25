package valueobject

import "testing"

func TestNewInteractionMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    InteractionMode
		wantErr bool
	}{
		{"collaboration", "collaboration", Collaboration, false},
		{"x-as-a-service", "x-as-a-service", XAsAService, false},
		{"facilitating", "facilitating", Facilitating, false},
		{"invalid", "invalid-mode", InteractionMode(""), true},
		{"empty string", "", InteractionMode(""), true},
		{"uppercase", "Collaboration", InteractionMode(""), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mode, err := NewInteractionMode(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if mode != tc.want {
					t.Errorf("NewInteractionMode(%q) = %q, want %q", tc.input, mode, tc.want)
				}
			}
		})
	}
}

func TestInteractionMode_String(t *testing.T) {
	tests := []struct {
		mode InteractionMode
		want string
	}{
		{Collaboration, "collaboration"},
		{XAsAService, "x-as-a-service"},
		{Facilitating, "facilitating"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if tc.mode.String() != tc.want {
				t.Errorf("String() = %q, want %q", tc.mode.String(), tc.want)
			}
		})
	}
}

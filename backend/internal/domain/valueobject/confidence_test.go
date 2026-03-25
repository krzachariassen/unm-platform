package valueobject

import "testing"

func TestNewConfidence(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		evidence string
		wantErr  bool
	}{
		{"score 0.0", 0.0, "no evidence", false},
		{"score 0.5", 0.5, "some evidence", false},
		{"score 1.0", 1.0, "strong evidence", false},
		{"score 0.75", 0.75, "evidence text", false},
		{"score -0.1", -0.1, "evidence", true},
		{"score 1.1", 1.1, "evidence", true},
		{"score -1.0", -1.0, "evidence", true},
		{"score 2.0", 2.0, "evidence", true},
		{"empty evidence is ok", 0.8, "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewConfidence(tc.score, tc.evidence)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for score %v, got nil", tc.score)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for score %v: %v", tc.score, err)
				}
				if c.Score != tc.score {
					t.Errorf("Score = %v, want %v", c.Score, tc.score)
				}
				if c.Evidence != tc.evidence {
					t.Errorf("Evidence = %q, want %q", c.Evidence, tc.evidence)
				}
			}
		})
	}
}

func TestConfidence_IsLow(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		isLow bool
	}{
		{"0.0 is low", 0.0, true},
		{"0.49 is low", 0.49, true},
		{"0.5 is not low", 0.5, false},
		{"0.51 is not low", 0.51, false},
		{"1.0 is not low", 1.0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewConfidence(tc.score, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.IsLow() != tc.isLow {
				t.Errorf("IsLow() = %v, want %v for score %v", c.IsLow(), tc.isLow, tc.score)
			}
		})
	}
}

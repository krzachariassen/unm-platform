package valueobject

import "fmt"

// Confidence represents an AI-inferred mapping confidence with a score and evidence.
type Confidence struct {
	Score    float64
	Evidence string
}

// NewConfidence constructs a Confidence value. Returns an error if score is outside [0.0, 1.0].
func NewConfidence(score float64, evidence string) (Confidence, error) {
	if score < 0.0 || score > 1.0 {
		return Confidence{}, fmt.Errorf("confidence: score %v is out of range [0.0, 1.0]", score)
	}
	return Confidence{Score: score, Evidence: evidence}, nil
}

// IsLow returns true if the confidence score is below 0.5.
func (c Confidence) IsLow() bool {
	return c.Score < 0.5
}

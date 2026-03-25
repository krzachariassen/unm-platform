package valueobject

import "fmt"

// Severity represents the severity level of a pain point or issue.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

var severityLevels = map[Severity]int{
	SeverityLow:      1,
	SeverityMedium:   2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

// NewSeverity constructs a Severity from a string. Returns an error if unrecognized.
func NewSeverity(s string) (Severity, error) {
	sv := Severity(s)
	if _, ok := severityLevels[sv]; !ok {
		return Severity(""), fmt.Errorf("severity: unrecognized value %q", s)
	}
	return sv, nil
}

// String returns the string representation of the Severity.
func (s Severity) String() string {
	return string(s)
}

// Level returns a numeric level for the severity (1=low, 2=medium, 3=high, 4=critical).
func (s Severity) Level() int {
	return severityLevels[s]
}

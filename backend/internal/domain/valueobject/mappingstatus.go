package valueobject

import "fmt"

// MappingStatus represents the assertion status of a relationship mapping.
type MappingStatus string

const (
	Asserted   MappingStatus = "asserted"
	Inferred   MappingStatus = "inferred"
	Candidate  MappingStatus = "candidate"
	Deprecated MappingStatus = "deprecated"
)

var validMappingStatuses = map[MappingStatus]bool{
	Asserted:   true,
	Inferred:   true,
	Candidate:  true,
	Deprecated: true,
}

// NewMappingStatus constructs a MappingStatus from a string. Returns an error if unrecognized.
func NewMappingStatus(s string) (MappingStatus, error) {
	ms := MappingStatus(s)
	if !validMappingStatuses[ms] {
		return MappingStatus(""), fmt.Errorf("mappingstatus: unrecognized value %q", s)
	}
	return ms, nil
}

// String returns the string representation of the MappingStatus.
func (ms MappingStatus) String() string {
	return string(ms)
}

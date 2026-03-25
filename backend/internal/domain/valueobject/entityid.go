package valueobject

import (
	"errors"
	"strings"
)

// EntityID is a value object wrapping a non-empty string identifier.
type EntityID struct {
	value string
}

// NewEntityID constructs an EntityID. Returns an error if s is empty or whitespace-only.
func NewEntityID(s string) (EntityID, error) {
	if strings.TrimSpace(s) == "" {
		return EntityID{}, errors.New("entityid: value must not be empty or whitespace")
	}
	return EntityID{value: s}, nil
}

// String returns the string representation of the EntityID.
func (e EntityID) String() string {
	return e.value
}

// IsZero returns true if the EntityID holds no value (zero value).
func (e EntityID) IsZero() bool {
	return e.value == ""
}

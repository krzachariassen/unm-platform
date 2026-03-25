package valueobject

import "fmt"

// InteractionMode represents how two teams interact with each other.
type InteractionMode string

const (
	Collaboration InteractionMode = "collaboration"
	XAsAService   InteractionMode = "x-as-a-service"
	Facilitating  InteractionMode = "facilitating"
)

var validInteractionModes = map[InteractionMode]bool{
	Collaboration: true,
	XAsAService:   true,
	Facilitating:  true,
}

// NewInteractionMode constructs an InteractionMode from a string. Returns an error if unrecognized.
func NewInteractionMode(s string) (InteractionMode, error) {
	mode := InteractionMode(s)
	if !validInteractionModes[mode] {
		return InteractionMode(""), fmt.Errorf("interactionmode: unrecognized value %q", s)
	}
	return mode, nil
}

// String returns the string representation of the InteractionMode.
func (m InteractionMode) String() string {
	return string(m)
}

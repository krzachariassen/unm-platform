package valueobject

import "fmt"

// TeamType represents the Team Topologies type of a team.
type TeamType string

const (
	StreamAligned        TeamType = "stream-aligned"
	Platform             TeamType = "platform"
	Enabling             TeamType = "enabling"
	ComplicatedSubsystem TeamType = "complicated-subsystem"
)

var validTeamTypes = map[TeamType]bool{
	StreamAligned:        true,
	Platform:             true,
	Enabling:             true,
	ComplicatedSubsystem: true,
}

// NewTeamType constructs a TeamType from a string. Returns an error if unrecognized.
func NewTeamType(s string) (TeamType, error) {
	tt := TeamType(s)
	if !tt.IsValid() {
		return TeamType(""), fmt.Errorf("teamtype: unrecognized value %q", s)
	}
	return tt, nil
}

// String returns the string representation of the TeamType.
func (tt TeamType) String() string {
	return string(tt)
}

// IsValid returns true if the TeamType is one of the recognized values.
func (tt TeamType) IsValid() bool {
	return validTeamTypes[tt]
}

package valueobject

import "fmt"

// RelationshipRole represents the role of a party in a relationship.
type RelationshipRole string

const (
	Primary    RelationshipRole = "primary"
	Supporting RelationshipRole = "supporting"
	Consuming  RelationshipRole = "consuming"
)

var validRelationshipRoles = map[RelationshipRole]bool{
	Primary:    true,
	Supporting: true,
	Consuming:  true,
}

// NewRelationshipRole constructs a RelationshipRole from a string.
// An empty string is valid and means "no role specified".
// Returns an error if the value is non-empty and unrecognized.
func NewRelationshipRole(s string) (RelationshipRole, error) {
	if s == "" {
		return RelationshipRole(""), nil
	}
	rr := RelationshipRole(s)
	if !validRelationshipRoles[rr] {
		return RelationshipRole(""), fmt.Errorf("relationshiprole: unrecognized value %q", s)
	}
	return rr, nil
}

// String returns the string representation of the RelationshipRole.
func (rr RelationshipRole) String() string {
	return string(rr)
}

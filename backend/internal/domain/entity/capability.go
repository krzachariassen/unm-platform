package entity

import (
	"errors"
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// Capability visibility constants.
const (
	CapVisibilityUserFacing     = "user-facing"
	CapVisibilityDomain         = "domain"
	CapVisibilityFoundational   = "foundational"
	CapVisibilityInfrastructure = "infrastructure"
)

var validCapabilityVisibilities = map[string]bool{
	CapVisibilityUserFacing:     true,
	CapVisibilityDomain:         true,
	CapVisibilityFoundational:   true,
	CapVisibilityInfrastructure: true,
	"":                          true, // empty is allowed (not required)
}

// Capability represents what the system or organisation must be able to do to support a need.
// Capabilities can nest hierarchically.
type Capability struct {
	ID          valueobject.EntityID
	Name        string
	Description string
	Visibility  string
	Children    []*Capability
	DependsOn   []Relationship
	// DecomposesTo is an alias for Children.
	DecomposesTo []*Capability
}

// NewCapability constructs a Capability. Returns an error if name is empty.
func NewCapability(id, name, description string) (*Capability, error) {
	if name == "" {
		return nil, errors.New("capability: name must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	c := &Capability{
		ID:          entityID,
		Name:        name,
		Description: description,
		Children:    []*Capability{},
		DependsOn:   []Relationship{},
	}
	c.DecomposesTo = c.Children
	return c, nil
}

// SetVisibility sets the capability visibility layer. Returns error if value is not valid.
func (c *Capability) SetVisibility(v string) error {
	if !validCapabilityVisibilities[v] {
		return fmt.Errorf("capability: invalid visibility %q", v)
	}
	c.Visibility = v
	return nil
}

// AddChild appends a child Capability.
func (c *Capability) AddChild(child *Capability) {
	c.Children = append(c.Children, child)
	c.DecomposesTo = c.Children
}

// AddDependsOn appends a Relationship to DependsOn.
func (c *Capability) AddDependsOn(r Relationship) {
	c.DependsOn = append(c.DependsOn, r)
}

// IsLeaf returns true if the Capability has no children.
func (c *Capability) IsLeaf() bool {
	return len(c.Children) == 0
}

// Depth returns the depth of the capability tree rooted at this node.
// A leaf node has depth 1.
func (c *Capability) Depth() int {
	if c.IsLeaf() {
		return 1
	}
	max := 0
	for _, child := range c.Children {
		d := child.Depth()
		if d > max {
			max = d
		}
	}
	return 1 + max
}

// IsFragmented returns true if more than 2 teams own this capability.
func (c *Capability) IsFragmented(teams []string) bool {
	return len(teams) > 2
}

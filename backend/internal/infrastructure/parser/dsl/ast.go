package dsl

import "fmt"

// ParseError is a structured parse error that includes source line information.
type ParseError struct {
	Line    int
	Message string
}

// Error implements the error interface. Format: "line N: message".
func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

// File is the root AST node representing a parsed .unm file.
type File struct {
	System               *SystemNode
	Actors               []*ActorNode
	Needs                []*NeedNode
	Capabilities         []*CapabilityNode
	Services             []*ServiceNode
	Teams                []*TeamNode
	Platforms            []*PlatformNode
	Interactions         []*InteractionNode
	DataAssets           []*DataAssetNode
	ExternalDependencies []*ExternalDependencyNode
	Signals              []*SignalNode
	Imports              []*ImportNode
	InferredMappings     []*InferredMappingNode
	Transitions          []*TransitionNode
}

// SystemNode represents a system block in the DSL.
type SystemNode struct {
	Name         string
	Description  string
	Version      string
	LastModified string
	Author       string
}

// ActorNode represents an actor block in the DSL.
type ActorNode struct {
	Name        string
	Description string
}

// NeedNode represents a need block in the DSL.
type NeedNode struct {
	Name        string
	Description string
	Outcome     string
	Actors      []string
	SupportedBy []RelationshipNode
}

// CapabilityNode represents a capability block in the DSL, which may nest children.
type CapabilityNode struct {
	Name        string
	Description string
	Visibility  string
	Parent      string // flat parent reference (9.1.2)
	DependsOn   []RelationshipNode
	Children    []*CapabilityNode
}

// RelationshipNode represents a directed reference with optional description and role.
type RelationshipNode struct {
	Target      string
	Description string
	Role        string // primary, supporting, consuming, or ""
}

// ServiceRealizesNode represents a service realizes relationship in the DSL.
type ServiceRealizesNode struct {
	Target string
	Role   string // primary, supporting, consuming, or ""
}

// ServiceNode represents a service block in the DSL.
type ServiceNode struct {
	Name         string
	Description  string
	OwnedBy      string
	DependsOn    []RelationshipNode
	Realizes     []ServiceRealizesNode // 9.3.5
	ExternalDeps []string              // 9.4.3
}

// TeamInteractionNode represents an inline interaction declaration inside a team block.
type TeamInteractionNode struct {
	With        string
	Mode        string
	Via         string
	Description string
}

// TeamNode represents a team block in the DSL.
type TeamNode struct {
	Name        string
	Type        string
	Description string
	Size        int
	Owns        []string
	Interacts   []TeamInteractionNode // 9.5.3
}

// PlatformNode represents a platform block in the DSL.
type PlatformNode struct {
	Name        string
	Description string
	Teams       []string
}

// InteractionNode represents an interaction block in the DSL.
type InteractionNode struct {
	From        string
	To          string
	Via         string
	Mode        string
	Description string
}

// DataAssetNode represents a data_asset block in the DSL.
type DataAssetNode struct {
	Name        string
	Type        string
	Description string
	UsedBy      []string
}

// ExternalDepUsageNode represents a service using an external dependency, with optional description.
type ExternalDepUsageNode struct {
	Target      string
	Description string
}

// ExternalDependencyNode represents an external_dependency block in the DSL.
type ExternalDependencyNode struct {
	Name        string
	Description string
	UsedBy      []ExternalDepUsageNode
}

// SignalNode represents a signal block in the DSL.
type SignalNode struct {
	Name        string
	Category    string
	Severity    string
	Description string
	OnEntity    string
	Affected    []string
}

// ImportNode represents an import statement in the DSL.
// Simple form:  import "path.unm"
// Named form:   import alias from "path.unm"
type ImportNode struct {
	Path  string // the file path in quotes
	Alias string // optional: "actors" in "import actors from path"
}

// TransitionNode represents a transition block in the DSL.
type TransitionNode struct {
	Name        string
	Description string
	Current     []TransitionBindingNode
	Target      []TransitionBindingNode
	Steps       []TransitionStepNode
}

// TransitionBindingNode represents a capability-team ownership pair in a transition state.
type TransitionBindingNode struct {
	CapabilityName string
	TeamName       string
}

// TransitionStepNode represents a step in a transition plan.
type TransitionStepNode struct {
	Number          int
	Label           string
	ActionText      string
	ExpectedOutcome string
}

// InferredMappingNode represents an inferred block in the DSL.
type InferredMappingNode struct {
	From       string  // need or entity name
	To         string  // capability name
	Confidence float64 // 0.0-1.0
	Evidence   string
	Status     string // suggested, confirmed, rejected
}

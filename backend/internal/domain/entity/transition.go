package entity

// Transition represents a planned evolution of the architecture model.
type Transition struct {
	Name        string
	Description string
	Current     []TransitionBinding
	Target      []TransitionBinding
	Steps       []TransitionStep
}

// TransitionBinding is a capability-team ownership pair in a state snapshot.
type TransitionBinding struct {
	CapabilityName string
	TeamName       string
}

// TransitionStep is a step in the transition plan.
type TransitionStep struct {
	Number          int
	Label           string
	ActionText      string
	ExpectedOutcome string
}

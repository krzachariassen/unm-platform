package entity

import (
	"fmt"
	"slices"
	"sort"
)

// System holds top-level system information.
type System struct {
	Name        string
	Description string
}

// ModelSummary provides a high-level summary of the UNMModel.
type ModelSummary struct {
	SystemName           string
	SystemDescription    string
	ActorCount           int
	NeedCount            int
	CapabilityCount      int
	ServiceCount         int
	TeamCount            int
	OrphanServiceCount   int
	FragmentedCapCount   int
	OverloadedTeamCount  int
}

// UNMModel is the root aggregate for a User Needs Map.
type UNMModel struct {
	System               System
	Actors               map[string]*Actor
	Needs                map[string]*Need
	Capabilities         map[string]*Capability
	CapabilityParents    map[string]string // child name → parent name
	Services             map[string]*Service
	Teams                map[string]*Team
	Platforms            map[string]*Platform
	Interactions         []*Interaction
	Signals              []*Signal
	DataAssets           map[string]*DataAsset
	ExternalDependencies map[string]*ExternalDependency
	InferredMappings     []*InferredMapping
	Transitions          []*Transition
	Warnings             []string // deprecation and reference warnings from parsing
}

// NewUNMModel constructs a UNMModel with all maps and slices initialized.
func NewUNMModel(systemName, systemDescription string) *UNMModel {
	return &UNMModel{
		System: System{
			Name:        systemName,
			Description: systemDescription,
		},
		Actors:               make(map[string]*Actor),
		Needs:                make(map[string]*Need),
		Capabilities:         make(map[string]*Capability),
		CapabilityParents:    make(map[string]string),
		Services:             make(map[string]*Service),
		Teams:                make(map[string]*Team),
		Platforms:            make(map[string]*Platform),
		Interactions:         []*Interaction{},
		Signals:              []*Signal{},
		DataAssets:           make(map[string]*DataAsset),
		ExternalDependencies: make(map[string]*ExternalDependency),
		InferredMappings:     []*InferredMapping{},
		Transitions:          []*Transition{},
	}
}

// AddActor adds an Actor to the model. Returns an error if a duplicate name exists.
func (m *UNMModel) AddActor(a *Actor) error {
	if _, exists := m.Actors[a.Name]; exists {
		return fmt.Errorf("model: actor %q already exists", a.Name)
	}
	m.Actors[a.Name] = a
	return nil
}

// AddNeed adds a Need to the model. Returns an error if a duplicate name exists.
func (m *UNMModel) AddNeed(n *Need) error {
	if _, exists := m.Needs[n.Name]; exists {
		return fmt.Errorf("model: need %q already exists", n.Name)
	}
	m.Needs[n.Name] = n
	return nil
}

// AddCapability adds a Capability and all its children recursively to the flat Capabilities map.
// Also populates CapabilityParents for each child registered.
// Returns an error if a duplicate name exists at any level.
func (m *UNMModel) AddCapability(c *Capability) error {
	return m.addCapabilityWithParent(c, "")
}

func (m *UNMModel) addCapabilityWithParent(c *Capability, parentName string) error {
	if _, exists := m.Capabilities[c.Name]; exists {
		return fmt.Errorf("model: capability %q already exists", c.Name)
	}
	m.Capabilities[c.Name] = c
	if parentName != "" {
		m.CapabilityParents[c.Name] = parentName
	}
	for _, child := range c.Children {
		if err := m.addCapabilityWithParent(child, c.Name); err != nil {
			return err
		}
	}
	return nil
}

// AddService adds a Service to the model. Returns an error if a duplicate name exists.
func (m *UNMModel) AddService(s *Service) error {
	if _, exists := m.Services[s.Name]; exists {
		return fmt.Errorf("model: service %q already exists", s.Name)
	}
	m.Services[s.Name] = s
	return nil
}

// AddTeam adds a Team to the model. Returns an error if a duplicate name exists.
func (m *UNMModel) AddTeam(t *Team) error {
	if _, exists := m.Teams[t.Name]; exists {
		return fmt.Errorf("model: team %q already exists", t.Name)
	}
	m.Teams[t.Name] = t
	return nil
}

// AddPlatform adds a Platform to the model. Returns an error if a duplicate name exists.
func (m *UNMModel) AddPlatform(p *Platform) error {
	if _, exists := m.Platforms[p.Name]; exists {
		return fmt.Errorf("model: platform %q already exists", p.Name)
	}
	m.Platforms[p.Name] = p
	return nil
}

// AddInteraction appends an Interaction (no duplicate check).
func (m *UNMModel) AddInteraction(i *Interaction) {
	m.Interactions = append(m.Interactions, i)
}

// AddSignal appends a Signal (no duplicate check).
func (m *UNMModel) AddSignal(s *Signal) {
	m.Signals = append(m.Signals, s)
}

// AddDataAsset adds a DataAsset to the model. Returns an error if a duplicate name exists.
func (m *UNMModel) AddDataAsset(d *DataAsset) error {
	if _, exists := m.DataAssets[d.Name]; exists {
		return fmt.Errorf("model: data asset %q already exists", d.Name)
	}
	m.DataAssets[d.Name] = d
	return nil
}

// AddExternalDependency adds an ExternalDependency. Returns an error if a duplicate name exists.
func (m *UNMModel) AddExternalDependency(e *ExternalDependency) error {
	if _, exists := m.ExternalDependencies[e.Name]; exists {
		return fmt.Errorf("model: external dependency %q already exists", e.Name)
	}
	m.ExternalDependencies[e.Name] = e
	return nil
}

// AddInferredMapping appends an InferredMapping (no duplicate check).
func (m *UNMModel) AddInferredMapping(im *InferredMapping) {
	m.InferredMappings = append(m.InferredMappings, im)
}

// AddTransition appends a Transition (no duplicate check).
func (m *UNMModel) AddTransition(t *Transition) {
	m.Transitions = append(m.Transitions, t)
}

// GetCapabilitiesForTeam returns all Capabilities owned by the named team.
// Ownership is determined by team.Owns[].TargetID matching the capability name.
func (m *UNMModel) GetCapabilitiesForTeam(teamName string) []*Capability {
	team, ok := m.Teams[teamName]
	if !ok {
		return nil
	}
	var result []*Capability
	for _, rel := range team.Owns {
		if cap, found := m.Capabilities[rel.TargetID.String()]; found {
			result = append(result, cap)
		}
	}
	return result
}

// GetServicesForCapability returns services referenced in cap.RealizedBy (top-down lookup).
func (m *UNMModel) GetServicesForCapability(capabilityName string) []*Service {
	cap, ok := m.Capabilities[capabilityName]
	if !ok {
		return nil
	}
	var result []*Service
	for _, rel := range cap.RealizedBy {
		if svc, found := m.Services[rel.TargetID.String()]; found {
			result = append(result, svc)
		}
	}
	return result
}

// GetCapabilitiesForService returns capabilities whose RealizedBy references the named service.
func (m *UNMModel) GetCapabilitiesForService(serviceName string) []*Capability {
	var result []*Capability
	for _, cap := range m.Capabilities {
		for _, rel := range cap.RealizedBy {
			if rel.TargetID.String() == serviceName {
				result = append(result, cap)
				break
			}
		}
	}
	return result
}

// GetTeamsForCapability returns all Teams that own the named capability.
func (m *UNMModel) GetTeamsForCapability(capabilityName string) []*Team {
	var result []*Team
	for _, team := range m.Teams {
		for _, rel := range team.Owns {
			if rel.TargetID.String() == capabilityName {
				result = append(result, team)
				break
			}
		}
	}
	return result
}

// GetNeedsForActor returns all Needs that have the given actor in their ActorNames.
func (m *UNMModel) GetNeedsForActor(actorName string) []*Need {
	var result []*Need
	for _, need := range m.Needs {
		if need.HasActor(actorName) {
			result = append(result, need)
		}
	}
	return result
}

// GetOrphanServices returns services not referenced in any capability's RealizedBy.
func (m *UNMModel) GetOrphanServices() []*Service {
	referenced := make(map[string]bool)
	for _, cap := range m.Capabilities {
		for _, rel := range cap.RealizedBy {
			referenced[rel.TargetID.String()] = true
		}
	}
	var result []*Service
	for name, svc := range m.Services {
		if !referenced[name] {
			result = append(result, svc)
		}
	}
	return result
}

// GetFragmentedCapabilities returns all Capabilities owned by more than 2 teams.
func (m *UNMModel) GetFragmentedCapabilities() []*Capability {
	var result []*Capability
	for name, cap := range m.Capabilities {
		teams := m.GetTeamsForCapability(name)
		if len(teams) > 2 {
			result = append(result, cap)
		}
	}
	return result
}

// GetOverloadedTeams returns all Teams where IsOverloaded() is true (capability count > threshold).
func (m *UNMModel) GetOverloadedTeams(threshold int) []*Team {
	var result []*Team
	for _, team := range m.Teams {
		if team.IsOverloaded(threshold) {
			result = append(result, team)
		}
	}
	return result
}

// GetRootCapabilities returns capabilities that have no parent (top-level only).
func (m *UNMModel) GetRootCapabilities() []*Capability {
	var result []*Capability
	for name, cap := range m.Capabilities {
		if _, hasParent := m.CapabilityParents[name]; !hasParent {
			result = append(result, cap)
		}
	}
	return result
}

// GetCapabilityPath returns the chain of capability names from root to the named capability.
// Returns []string{name} if root. Returns nil if not found.
func (m *UNMModel) GetCapabilityPath(capName string) []string {
	if _, exists := m.Capabilities[capName]; !exists {
		return nil
	}
	var path []string
	current := capName
	for {
		path = append([]string{current}, path...)
		parent, hasParent := m.CapabilityParents[current]
		if !hasParent {
			break
		}
		current = parent
	}
	return path
}

// GetCapabilitiesByLayer returns all capabilities with the given visibility value.
func (m *UNMModel) GetCapabilitiesByLayer(layer string) []*Capability {
	var result []*Capability
	for _, cap := range m.Capabilities {
		if cap.Visibility == layer {
			result = append(result, cap)
		}
	}
	return result
}

// GetDataAssetsForService returns DataAssets used by the named service.
func (m *UNMModel) GetDataAssetsForService(serviceName string) []*DataAsset {
	var result []*DataAsset
	for _, da := range m.DataAssets {
		if slices.Contains(da.UsedBy, serviceName) {
			result = append(result, da)
		}
	}
	return result
}

// GetExternalDepsForService returns ExternalDependencies used by the named service.
func (m *UNMModel) GetExternalDepsForService(serviceName string) []*ExternalDependency {
	var result []*ExternalDependency
	for _, ext := range m.ExternalDependencies {
		for _, u := range ext.UsedBy {
			if u.ServiceName == serviceName {
				result = append(result, ext)
				break
			}
		}
	}
	return result
}

// GetPlatformForTeam returns the Platform that includes the named team, or nil if none.
func (m *UNMModel) GetPlatformForTeam(teamName string) *Platform {
	for _, p := range m.Platforms {
		if slices.Contains(p.TeamNames, teamName) {
			return p
		}
	}
	return nil
}

// Summary returns a ModelSummary with counts for each entity type and key metrics.
func (m *UNMModel) Summary() ModelSummary {
	orphans := m.GetOrphanServices()
	fragmented := m.GetFragmentedCapabilities()
	overloaded := m.GetOverloadedTeams(DefaultConfig().Analysis.OverloadedCapabilityThreshold)

	return ModelSummary{
		SystemName:           m.System.Name,
		SystemDescription:    m.System.Description,
		ActorCount:           len(m.Actors),
		NeedCount:            len(m.Needs),
		CapabilityCount:      len(m.Capabilities),
		ServiceCount:         len(m.Services),
		TeamCount:            len(m.Teams),
		OrphanServiceCount:   len(orphans),
		FragmentedCapCount:   len(fragmented),
		OverloadedTeamCount:  len(overloaded),
	}
}

// ValueChainLayer groups capabilities at the same visibility level.
type ValueChainLayer struct {
	Layer        string // e.g. "user-facing", "domain", "foundational", "infrastructure", ""
	Capabilities []*Capability
}

// UNMMapEntry represents one path: actor → need → capability chain.
type UNMMapEntry struct {
	ActorName    string
	NeedName     string
	Capabilities []*Capability // capabilities supporting this need
}

// canonicalLayerOrder defines the display order for value chain layers.
var canonicalLayerOrder = []string{
	CapVisibilityUserFacing,
	CapVisibilityDomain,
	CapVisibilityFoundational,
	CapVisibilityInfrastructure,
	"",
}

// BuildValueChain returns all capabilities grouped by visibility layer.
// Order: user-facing, domain, foundational, infrastructure, then any unset ("").
// Within each layer, capabilities are sorted by name for determinism.
// Layers with no capabilities are omitted.
func (m *UNMModel) BuildValueChain() []ValueChainLayer {
	// Build a map from layer → capabilities.
	byLayer := make(map[string][]*Capability)
	for _, cap := range m.Capabilities {
		byLayer[cap.Visibility] = append(byLayer[cap.Visibility], cap)
	}

	var result []ValueChainLayer
	for _, layer := range canonicalLayerOrder {
		caps, ok := byLayer[layer]
		if !ok || len(caps) == 0 {
			continue
		}
		sort.Slice(caps, func(i, j int) bool {
			return caps[i].Name < caps[j].Name
		})
		result = append(result, ValueChainLayer{
			Layer:        layer,
			Capabilities: caps,
		})
	}
	return result
}

// BuildUNMMap returns the actor→need→capability chain for all actors (or filtered by actorFilter).
// If actorFilter is empty, include all actors.
// Each Need becomes one UNMMapEntry per actor it belongs to.
// Entries are sorted by ActorName, then NeedName for determinism.
func (m *UNMModel) BuildUNMMap(actorFilter string) []UNMMapEntry {
	var entries []UNMMapEntry
	for _, need := range m.Needs {
		var caps []*Capability
		for _, rel := range need.SupportedBy {
			if cap, found := m.Capabilities[rel.TargetID.String()]; found {
				caps = append(caps, cap)
			}
		}
		for _, actorName := range need.ActorNames {
			if actorFilter != "" && actorName != actorFilter {
				continue
			}
			entries = append(entries, UNMMapEntry{
				ActorName:    actorName,
				NeedName:     need.Name,
				Capabilities: caps,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ActorName != entries[j].ActorName {
			return entries[i].ActorName < entries[j].ActorName
		}
		return entries[i].NeedName < entries[j].NeedName
	})
	return entries
}

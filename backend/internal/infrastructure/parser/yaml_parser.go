package parser

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// Parser defines the interface for parsing a UNM model from an io.Reader.
type Parser interface {
	Parse(r io.Reader) (*entity.UNMModel, error)
}

// YAMLParser implements Parser for YAML-encoded UNM models.
type YAMLParser struct{}

// NewYAMLParser constructs a YAMLParser.
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// ParseFile is a convenience function that opens a file and parses it as a UNM YAML model.
func ParseFile(path string) (*entity.UNMModel, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parser: open file %q: %w", path, err)
	}
	defer f.Close()
	return NewYAMLParser().Parse(f)
}

// Parse reads YAML from r and returns a fully constructed UNMModel.
func (p *YAMLParser) Parse(r io.Reader) (*entity.UNMModel, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("parser: read input: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("parser: input is empty")
	}

	var raw yamlDocument
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parser: invalid YAML: %w", err)
	}

	return buildModel(&raw)
}

// ---------------------------------------------------------------------------
// YAML schema structs
// ---------------------------------------------------------------------------

type yamlDocument struct {
	System yamlSystem  `yaml:"system"`
	Actors []yamlActor `yaml:"actors"`
	// Scenarios section is ignored (deprecated in 1.9.1) — field kept for backward-compat YAML parsing.
	Scenarios            []yamlScenario           `yaml:"scenarios"`
	Needs                []yamlNeed               `yaml:"needs"`
	Capabilities         []yamlCapability         `yaml:"capabilities"`
	Services             []yamlService            `yaml:"services"`
	Teams                []yamlTeam               `yaml:"teams"`
	Platforms            []yamlPlatform           `yaml:"platforms"`
	Interactions         []yamlInteraction        `yaml:"interactions"`
	// Signals, PainPoints, and Inferred sections are ignored (removed in DSL v2.0).
	// These are now computed by the platform's analyzers, not user-authored.
	Signals    []map[string]any   `yaml:"signals"`
	PainPoints []map[string]any   `yaml:"pain_points"`
	Inferred   []map[string]any   `yaml:"inferred"`
	DataAssets           []yamlDataAssetFlex      `yaml:"data_assets"`
	ExternalDependencies []yamlExternalDependency `yaml:"external_dependencies"`
}

type yamlSystem struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type yamlActor struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// yamlScenario is kept for backward-compatible YAML parsing but scenarios are not imported into the model.
type yamlScenario struct {
	Name        string `yaml:"name"`
	Actor       string `yaml:"actor"`
	Description string `yaml:"description"`
}

type yamlNeed struct {
	Name  string       `yaml:"name"`
	Actor stringOrList `yaml:"actor"`
	// Scenario field is ignored (deprecated in 1.9.1) — kept for backward-compat YAML parsing.
	Scenario    string             `yaml:"scenario"`
	Outcome     string             `yaml:"outcome"`
	SupportedBy []flexRelationship `yaml:"supportedBy"`
}

// stringOrList accepts either a YAML scalar string or a sequence of strings.
type stringOrList struct {
	values []string
}

func (s *stringOrList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.SequenceNode {
		return value.Decode(&s.values)
	}
	var single string
	if err := value.Decode(&single); err != nil {
		return err
	}
	s.values = []string{single}
	return nil
}

type yamlCapability struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Visibility  string             `yaml:"visibility"`
	Parent      string             `yaml:"parent"`      // 9.1: flat form parent reference
	OwnedBy     string             `yaml:"ownedBy"`     // 9.6: deprecated, triggers warning
	Children    []yamlCapability   `yaml:"children"`
	RealizedBy  []flexRelationship `yaml:"realizedBy"`
	DependsOn   []flexRelationship `yaml:"dependsOn"`
}

type yamlService struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	OwnedBy     string `yaml:"ownedBy"`
	// Supports, DataAssets, ExternalDependsOn are deprecated (1.9.4) and silently ignored.
	// TODO: emit deprecation warnings in a future version.
	Supports          []flexRelationship `yaml:"supports"`
	DependsOn         []flexRelationship `yaml:"dependsOn"`
	ExternalDependsOn []flexRelationship `yaml:"externalDependsOn"`
	DataAssets        []yamlDataAssetRef `yaml:"dataAssets"`
	Realizes          []flexRelationship `yaml:"realizes"`    // 9.3: service declares what it realizes
	ExternalDeps      []string           `yaml:"externalDeps"` // 9.4: service declares its external deps
}

// flexDataAssetUsedBy handles both compact map form and object list form for data_asset.usedBy
type flexDataAssetUsedBy struct {
	entries []entity.DataAssetServiceUsage
}

func (f *flexDataAssetUsedBy) UnmarshalYAML(value *yaml.Node) error {
	// Map form: {svc-a: "read-write", svc-b: "read"}
	if value.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(value.Content); i += 2 {
			key := value.Content[i].Value
			val := value.Content[i+1].Value
			f.entries = append(f.entries, entity.DataAssetServiceUsage{
				ServiceName: key,
				Access:      val,
			})
		}
		return nil
	}
	// Sequence form: [{target: svc, access: rw}, ...]
	if value.Kind == yaml.SequenceNode {
		for _, item := range value.Content {
			var ref yamlDataAssetRef
			if err := item.Decode(&ref); err != nil {
				return err
			}
			f.entries = append(f.entries, entity.DataAssetServiceUsage{
				ServiceName: ref.Target,
				Access:      ref.Access,
			})
		}
		return nil
	}
	return fmt.Errorf("parser: data_asset.usedBy must be a map or sequence")
}

type yamlDataAssetRef struct {
	Target string `yaml:"target"`
	Access string `yaml:"access"`
}

// yamlDataAssetFlex wraps yamlDataAsset to use flexDataAssetUsedBy
type yamlDataAssetFlex struct {
	Name        string              `yaml:"name"`
	Type        string              `yaml:"type"`
	Description string              `yaml:"description"`
	UsedBy      flexDataAssetUsedBy `yaml:"usedBy"`
	ProducedBy  string              `yaml:"producedBy"`
	ConsumedBy  []string            `yaml:"consumedBy"`
}

type yamlTeam struct {
	Name        string                `yaml:"name"`
	Type        string                `yaml:"type"`
	Description string                `yaml:"description"`
	Size        int                   `yaml:"size"`
	Owns        []flexRelationship    `yaml:"owns"`
	Interacts   []yamlTeamInteraction `yaml:"interacts"` // 9.5: inline interactions
}

// yamlTeamInteraction is the inline interaction form on a team.
type yamlTeamInteraction struct {
	With        string `yaml:"with"`
	Mode        string `yaml:"mode"`
	Via         string `yaml:"via"`
	Description string `yaml:"description"`
}

type yamlPlatform struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Teams       []string `yaml:"teams"`
}

type yamlInteraction struct {
	From        string `yaml:"from"`
	To          string `yaml:"to"`
	Mode        string `yaml:"mode"`
	Via         string `yaml:"via"`
	Description string `yaml:"description"`
}

type yamlExternalDependency struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	UsedBy      []flexRelationship `yaml:"usedBy"`
}

// ---------------------------------------------------------------------------
// flexRelationship — handles both short (string) and long (object) YAML forms
// ---------------------------------------------------------------------------

type yamlRelationship struct {
	Target      string `yaml:"target"`
	Description string `yaml:"description"`
	Role        string `yaml:"role"`
}

type flexRelationship struct {
	yamlRelationship
}

func (f *flexRelationship) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		f.Target = value.Value
		return nil
	}
	return value.Decode(&f.yamlRelationship)
}

// ---------------------------------------------------------------------------
// Model builder
// ---------------------------------------------------------------------------

func buildModel(raw *yamlDocument) (*entity.UNMModel, error) {
	if raw.System.Name == "" {
		return nil, fmt.Errorf("parser: system.name is required")
	}

	model := entity.NewUNMModel(raw.System.Name, raw.System.Description)
	var warnings []string

	if err := addActors(model, raw.Actors); err != nil {
		return nil, err
	}
	// raw.Scenarios is silently ignored (deprecated in 1.9.1).
	if err := addNeeds(model, raw.Needs); err != nil {
		return nil, err
	}
	capWarnings, err := addCapabilities(model, raw.Capabilities)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, capWarnings...)

	if err := addServices(model, raw.Services); err != nil {
		return nil, err
	}
	if err := addTeams(model, raw.Teams); err != nil {
		return nil, err
	}
	if err := addPlatforms(model, raw.Platforms); err != nil {
		return nil, err
	}

	// 9.5: Process inline team interactions before legacy interactions
	teamInterWarnings, err := addTeamInlineInteractions(model, raw.Teams)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, teamInterWarnings...)

	// 9.5: Legacy interactions section — emit deprecation warning if non-empty
	if len(raw.Interactions) > 0 {
		warnings = append(warnings, "parser: top-level interactions: section is deprecated; use team.interacts instead")
	}
	if err := addInteractions(model, raw.Interactions, &warnings); err != nil {
		return nil, err
	}

	// 9.8: Data assets with flex usedBy
	dataWarnings, err := addDataAssetsFromFlex(model, raw.DataAssets)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, dataWarnings...)

	// 9.4: External dependencies — detect legacy usedBy
	extWarnings, err := addExternalDependencies(model, raw.ExternalDependencies)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, extWarnings...)

	// 9.3: Post-process service.realizes — wire into capability.RealizedBy
	realizeWarnings, err := processServiceRealizes(model, raw.Services)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, realizeWarnings...)

	// 9.4: Post-process service.externalDeps — wire into ExternalDependency.UsedBy
	extDepWarnings, err := processServiceExternalDeps(model, raw.Services)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, extDepWarnings...)

	// 9.7: Reference validation
	refWarnings := validateReferences(model)
	warnings = append(warnings, refWarnings...)

	model.Warnings = warnings
	return model, nil
}

func addActors(model *entity.UNMModel, actors []yamlActor) error {
	for _, a := range actors {
		if a.Name == "" {
			return fmt.Errorf("parser: actor.name is required")
		}
		actor, err := entity.NewActor(a.Name, a.Name, a.Description)
		if err != nil {
			return fmt.Errorf("parser: actor %q: %w", a.Name, err)
		}
		if err := model.AddActor(&actor); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
}

func addNeeds(model *entity.UNMModel, needs []yamlNeed) error {
	for _, n := range needs {
		if n.Name == "" {
			return fmt.Errorf("parser: need.name is required")
		}
		actors := n.Actor.values
		if len(actors) == 0 {
			return fmt.Errorf("parser: need %q: need.actor is required", n.Name)
		}
		// n.Scenario is silently ignored (deprecated in 1.9.1).
		var need *entity.Need
		var err error
		if len(actors) == 1 {
			need, err = entity.NewNeed(n.Name, n.Name, actors[0], n.Outcome)
		} else {
			need, err = entity.NewNeedMultiActor(n.Name, n.Name, actors, n.Outcome)
		}
		if err != nil {
			return fmt.Errorf("parser: need %q: %w", n.Name, err)
		}
		for _, rel := range n.SupportedBy {
			r, err := buildRelationship(rel)
			if err != nil {
				return fmt.Errorf("parser: need %q supportedBy: %w", n.Name, err)
			}
			need.AddSupportedBy(r)
		}
		if err := model.AddNeed(need); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
}

// addCapabilities implements a two-pass algorithm:
// Pass 1: Build all capabilities (nested children via buildCapability).
// Pass 2: Resolve flat `parent` references and add as children.
// After hierarchy is built, apply visibility inheritance.
func addCapabilities(model *entity.UNMModel, caps []yamlCapability) ([]string, error) {
	var warnings []string

	// Pass 1: Build all capabilities and add to model.
	// buildCapability handles nested children recursively.
	for _, c := range caps {
		w, cap, err := buildCapabilityWithWarnings(c)
		if err != nil {
			return nil, err
		}
		warnings = append(warnings, w...)
		if err := model.AddCapability(cap); err != nil {
			return nil, fmt.Errorf("parser: %w", err)
		}
	}

	// Pass 2: Resolve flat `parent` references.
	// These are caps declared at the top level with a `parent:` field.
	for _, c := range caps {
		if c.Parent == "" {
			continue
		}
		// Check for circular: does the parent (directly or transitively) reference c?
		if err := detectCircular(c.Name, c.Parent, caps); err != nil {
			return nil, err
		}
		parentCap, ok := model.Capabilities[c.Parent]
		if !ok {
			return nil, fmt.Errorf("parser: capability %q references unknown parent %q", c.Name, c.Parent)
		}
		childCap, ok := model.Capabilities[c.Name]
		if !ok {
			return nil, fmt.Errorf("parser: capability %q not found in model (internal error)", c.Name)
		}
		parentCap.AddChild(childCap)
		model.CapabilityParents[c.Name] = c.Parent
	}

	// Apply visibility inheritance (depth-first, parent-first).
	applyVisibilityInheritance(model)

	return warnings, nil
}

// detectCircular checks whether childName's parent chain contains childName itself.
// caps is the full list of yamlCapability definitions (used for lookup by name).
func detectCircular(childName, parentName string, caps []yamlCapability) error {
	visited := map[string]bool{childName: true}
	current := parentName
	for current != "" {
		if visited[current] {
			return fmt.Errorf("parser: circular parent reference detected involving capability %q", current)
		}
		visited[current] = true
		// Find current's parent
		next := ""
		for _, c := range caps {
			if c.Name == current {
				next = c.Parent
				break
			}
		}
		current = next
	}
	return nil
}

// applyVisibilityInheritance does a parent-first DFS across the capability tree,
// inheriting parent visibility into children that have no visibility set.
func applyVisibilityInheritance(model *entity.UNMModel) {
	// Start with root capabilities (no parent)
	for capName, cap := range model.Capabilities {
		if _, hasParent := model.CapabilityParents[capName]; !hasParent {
			inheritVisibility(cap, cap.Visibility)
		}
	}
}

func inheritVisibility(cap *entity.Capability, parentVisibility string) {
	if cap.Visibility == "" && parentVisibility != "" {
		_ = cap.SetVisibility(parentVisibility)
	}
	for _, child := range cap.Children {
		effectiveVis := cap.Visibility
		inheritVisibility(child, effectiveVis)
	}
}

func buildCapabilityWithWarnings(yc yamlCapability) ([]string, *entity.Capability, error) {
	var warnings []string

	if yc.Name == "" {
		return nil, nil, fmt.Errorf("parser: capability.name is required")
	}
	cap, err := entity.NewCapability(yc.Name, yc.Name, yc.Description)
	if err != nil {
		return nil, nil, fmt.Errorf("parser: capability %q: %w", yc.Name, err)
	}

	if yc.Visibility != "" {
		if err := cap.SetVisibility(yc.Visibility); err != nil {
			return nil, nil, fmt.Errorf("parser: capability %q: %w", yc.Name, err)
		}
	}

	// 9.6: ownedBy deprecation warning
	if yc.OwnedBy != "" {
		warnings = append(warnings, fmt.Sprintf("parser: capability %q uses deprecated ownedBy field; use team.owns instead", yc.Name))
	}

	// 9.3: realizedBy deprecation warning
	if len(yc.RealizedBy) > 0 {
		warnings = append(warnings, fmt.Sprintf("parser: capability %q uses deprecated realizedBy field; use service.realizes instead", yc.Name))
	}

	// Build children first (bottom-up for nested form)
	for _, child := range yc.Children {
		childWarnings, childCap, err := buildCapabilityWithWarnings(child)
		if err != nil {
			return nil, nil, err
		}
		warnings = append(warnings, childWarnings...)
		cap.AddChild(childCap)
	}

	for _, rel := range yc.RealizedBy {
		r, err := buildRelationship(rel)
		if err != nil {
			return nil, nil, fmt.Errorf("parser: capability %q realizedBy: %w", yc.Name, err)
		}
		cap.AddRealizedBy(r)
	}

	for _, rel := range yc.DependsOn {
		r, err := buildRelationship(rel)
		if err != nil {
			return nil, nil, fmt.Errorf("parser: capability %q dependsOn: %w", yc.Name, err)
		}
		cap.AddDependsOn(r)
	}

	return warnings, cap, nil
}

func addServices(model *entity.UNMModel, services []yamlService) error {
	for _, s := range services {
		if s.Name == "" {
			return fmt.Errorf("parser: service.name is required")
		}
		// s.Type is silently ignored (deprecated in 1.10) — kept in schema for backward compat.
		svc, err := entity.NewService(s.Name, s.Name, s.Description, s.OwnedBy)
		if err != nil {
			return fmt.Errorf("parser: service %q: %w", s.Name, err)
		}
		// s.Supports, s.DataAssets, s.ExternalDependsOn are silently ignored (deprecated in 1.9.4).
		for _, rel := range s.DependsOn {
			r, err := buildRelationship(rel)
			if err != nil {
				return fmt.Errorf("parser: service %q dependsOn: %w", s.Name, err)
			}
			svc.AddDependsOn(r)
		}
		if err := model.AddService(svc); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
}

// processServiceRealizes wires service.realizes into capability.RealizedBy.
// Returns warnings for conflicts with existing capability.realizedBy declarations.
func processServiceRealizes(model *entity.UNMModel, services []yamlService) ([]string, error) {
	var warnings []string
	for _, s := range services {
		for _, rel := range s.Realizes {
			capName := rel.Target
			cap, ok := model.Capabilities[capName]
			if !ok {
				// Will be caught by reference validation (9.7), skip silently here
				continue
			}
			// Check for duplicate with existing realizedBy entries
			duplicate := false
			for _, existing := range cap.RealizedBy {
				if existing.TargetID.String() == s.Name {
					duplicate = true
					break
				}
			}
			if duplicate {
				warnings = append(warnings, fmt.Sprintf(
					"parser: duplicate realizes relationship between service %q and capability %q (declared in both service.realizes and capability.realizedBy); keeping one",
					s.Name, capName,
				))
				continue
			}
			// Build the relationship with service name as target
			targetID, err := valueobject.NewEntityID(s.Name)
			if err != nil {
				return nil, fmt.Errorf("parser: service %q realizes: %w", s.Name, err)
			}
			role, err := valueobject.NewRelationshipRole(rel.Role)
			if err != nil {
				return nil, fmt.Errorf("parser: service %q realizes role: %w", s.Name, err)
			}
			r := entity.NewRelationship(targetID, rel.Description, role)
			cap.AddRealizedBy(r)
		}
	}
	return warnings, nil
}

// processServiceExternalDeps wires service.externalDeps into ExternalDependency.UsedBy.
// Returns warnings for conflicts with legacy usedBy declarations.
func processServiceExternalDeps(model *entity.UNMModel, services []yamlService) ([]string, error) {
	var warnings []string
	for _, s := range services {
		for _, depName := range s.ExternalDeps {
			ext, ok := model.ExternalDependencies[depName]
			if !ok {
				// Will be caught by reference validation
				continue
			}
			// Check for duplicate
			duplicate := false
			for _, existing := range ext.UsedBy {
				if existing.ServiceName == s.Name {
					duplicate = true
					break
				}
			}
			if duplicate {
				warnings = append(warnings, fmt.Sprintf(
					"parser: duplicate external dependency relationship between service %q and %q (declared in both service.externalDeps and external_dependencies[].usedBy); keeping one",
					s.Name, depName,
				))
				continue
			}
			ext.AddUsedBy(s.Name, "")
		}
	}
	return warnings, nil
}

func addTeams(model *entity.UNMModel, teams []yamlTeam) error {
	for _, t := range teams {
		if t.Name == "" {
			return fmt.Errorf("parser: team.name is required")
		}
		teamType, err := valueobject.NewTeamType(t.Type)
		if err != nil {
			return fmt.Errorf("parser: team %q type: %w", t.Name, err)
		}
		team, err := entity.NewTeam(t.Name, t.Name, t.Description, teamType)
		if err != nil {
			return fmt.Errorf("parser: team %q: %w", t.Name, err)
		}
		if t.Size > 0 {
			team.Size = t.Size
			team.SizeExplicit = true
		}
		for _, rel := range t.Owns {
			r, err := buildRelationship(rel)
			if err != nil {
				return fmt.Errorf("parser: team %q owns: %w", t.Name, err)
			}
			team.AddOwns(r)
		}
		if err := model.AddTeam(team); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
}

// addTeamInlineInteractions processes team.interacts and adds them to model.Interactions.
// Returns warnings for deduplication with legacy interactions.
func addTeamInlineInteractions(model *entity.UNMModel, teams []yamlTeam) ([]string, error) {
	var warnings []string
	for _, t := range teams {
		for _, inter := range t.Interacts {
			mode, err := valueobject.NewInteractionMode(inter.Mode)
			if err != nil {
				return nil, fmt.Errorf("parser: team %q interacts with %q: mode: %w", t.Name, inter.With, err)
			}
			id := fmt.Sprintf("%s->%s->%s", t.Name, inter.With, inter.Via)
			interaction, err := entity.NewInteraction(id, t.Name, inter.With, mode, inter.Via, inter.Description)
			if err != nil {
				return nil, fmt.Errorf("parser: team %q interacts: %w", t.Name, err)
			}
			model.AddInteraction(interaction)
		}
	}
	return warnings, nil
}

func addPlatforms(model *entity.UNMModel, platforms []yamlPlatform) error {
	for _, p := range platforms {
		if p.Name == "" {
			return fmt.Errorf("parser: platform.name is required")
		}
		platform, err := entity.NewPlatform(p.Name, p.Name, p.Description)
		if err != nil {
			return fmt.Errorf("parser: platform %q: %w", p.Name, err)
		}
		for _, teamName := range p.Teams {
			platform.AddTeam(teamName)
		}
		if err := model.AddPlatform(platform); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
}

// interactionKey returns a deduplification key for an interaction.
func interactionKey(from, to, mode string) string {
	return fmt.Sprintf("%s->%s->%s", from, to, mode)
}

// addInteractions adds legacy top-level interactions to the model.
// Deduplicates against interactions already added from team.interacts only.
// Does NOT deduplicate within the legacy list itself (preserves backward compat).
func addInteractions(model *entity.UNMModel, interactions []yamlInteraction, warnings *[]string) error {
	// Build a set of existing interaction keys from team.interacts that were already added.
	teamInteractKeys := make(map[string]bool)
	for _, inter := range model.Interactions {
		teamInteractKeys[interactionKey(inter.FromTeamName, inter.ToTeamName, inter.Mode.String())] = true
	}

	for i, inter := range interactions {
		key := interactionKey(inter.From, inter.To, inter.Mode)
		mode, err := valueobject.NewInteractionMode(inter.Mode)
		if err != nil {
			return fmt.Errorf("parser: interaction[%d] mode: %w", i, err)
		}
		if teamInteractKeys[key] {
			*warnings = append(*warnings, fmt.Sprintf(
				"parser: duplicate interaction from %q to %q with mode %q (already declared in team.interacts); keeping one",
				inter.From, inter.To, inter.Mode,
			))
			continue
		}
		id := fmt.Sprintf("%s->%s->%s", inter.From, inter.To, inter.Via)
		interaction, err := entity.NewInteraction(id, inter.From, inter.To, mode, inter.Via, inter.Description)
		if err != nil {
			return fmt.Errorf("parser: interaction[%d]: %w", i, err)
		}
		model.AddInteraction(interaction)
	}
	return nil
}

// addDataAssetsFromFlex handles the full parse of data_assets including compact usedBy syntax.
func addDataAssetsFromFlex(model *entity.UNMModel, assets []yamlDataAssetFlex) ([]string, error) {
	var warnings []string
	for _, a := range assets {
		if a.Name == "" {
			return nil, fmt.Errorf("parser: data_asset.name is required")
		}
		da, err := entity.NewDataAsset(a.Name, a.Name, a.Type, a.Description)
		if err != nil {
			return nil, fmt.Errorf("parser: data_asset %q: %w", a.Name, err)
		}
		for _, u := range a.UsedBy.entries {
			da.AddUsedBy(u.ServiceName, u.Access)
		}
		da.ProducedBy = a.ProducedBy
		for _, c := range a.ConsumedBy {
			da.ConsumedBy = append(da.ConsumedBy, c)
		}
		if err := model.AddDataAsset(da); err != nil {
			return nil, fmt.Errorf("parser: %w", err)
		}
	}
	return warnings, nil
}

func addExternalDependencies(model *entity.UNMModel, deps []yamlExternalDependency) ([]string, error) {
	var warnings []string
	for _, d := range deps {
		if d.Name == "" {
			return nil, fmt.Errorf("parser: external_dependency.name is required")
		}
		ext, err := entity.NewExternalDependency(d.Name, d.Name, d.Description)
		if err != nil {
			return nil, fmt.Errorf("parser: external_dependency %q: %w", d.Name, err)
		}
		// 9.4: legacy usedBy — emit deprecation warning
		if len(d.UsedBy) > 0 {
			warnings = append(warnings, fmt.Sprintf(
				"parser: external_dependency %q uses deprecated usedBy field; use service.externalDeps instead",
				d.Name,
			))
			for _, u := range d.UsedBy {
				ext.AddUsedBy(u.Target, u.Description)
			}
		}
		if err := model.AddExternalDependency(ext); err != nil {
			return nil, fmt.Errorf("parser: %w", err)
		}
	}
	return warnings, nil
}

// validateReferences checks model integrity after full build and returns warnings
// for unresolved references (9.7). Does NOT return errors.
func validateReferences(model *entity.UNMModel) []string {
	var warnings []string

	// need.supportedBy → capabilities
	for needName, need := range model.Needs {
		for _, rel := range need.SupportedBy {
			target := rel.TargetID.String()
			if _, ok := model.Capabilities[target]; !ok {
				warnings = append(warnings, fmt.Sprintf(
					"parser: unresolved reference: need %q supportedBy %q — capability not found",
					needName, target,
				))
			}
		}
	}

	// capability.realizedBy → services
	for capName, cap := range model.Capabilities {
		for _, rel := range cap.RealizedBy {
			target := rel.TargetID.String()
			if _, ok := model.Services[target]; !ok {
				warnings = append(warnings, fmt.Sprintf(
					"parser: unresolved reference: capability %q realizedBy %q — service not found",
					capName, target,
				))
			}
		}
	}

	// service.dependsOn → services
	for svcName, svc := range model.Services {
		for _, rel := range svc.DependsOn {
			target := rel.TargetID.String()
			if _, ok := model.Services[target]; !ok {
				warnings = append(warnings, fmt.Sprintf(
					"parser: unresolved reference: service %q dependsOn %q — service not found",
					svcName, target,
				))
			}
		}
	}

	// team.owns → capabilities
	for teamName, team := range model.Teams {
		for _, rel := range team.Owns {
			target := rel.TargetID.String()
			if _, ok := model.Capabilities[target]; !ok {
				warnings = append(warnings, fmt.Sprintf(
					"parser: unresolved reference: team %q owns %q — capability not found",
					teamName, target,
				))
			}
		}
	}

	return warnings
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func buildRelationship(rel flexRelationship) (entity.Relationship, error) {
	targetID, err := valueobject.NewEntityID(rel.Target)
	if err != nil {
		return entity.Relationship{}, fmt.Errorf("relationship target: %w", err)
	}
	role, err := valueobject.NewRelationshipRole(rel.Role)
	if err != nil {
		return entity.Relationship{}, fmt.Errorf("relationship role: %w", err)
	}
	return entity.NewRelationship(targetID, rel.Description, role), nil
}

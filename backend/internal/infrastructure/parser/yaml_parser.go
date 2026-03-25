package parser

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
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
	Signals    []map[string]any `yaml:"signals"`
	PainPoints []map[string]any `yaml:"pain_points"`
	Inferred   []map[string]any `yaml:"inferred"`
	DataAssets           []yamlDataAsset          `yaml:"data_assets"`
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
	Name  string `yaml:"name"`
	Actor string `yaml:"actor"`
	// Scenario field is ignored (deprecated in 1.9.1) — kept for backward-compat YAML parsing.
	Scenario    string             `yaml:"scenario"`
	Outcome     string             `yaml:"outcome"`
	SupportedBy []flexRelationship `yaml:"supportedBy"`
}

type yamlCapability struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Visibility  string             `yaml:"visibility"`
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
}

type yamlDataAssetRef struct {
	Target string `yaml:"target"`
	Access string `yaml:"access"`
}

type yamlTeam struct {
	Name        string             `yaml:"name"`
	Type        string             `yaml:"type"`
	Description string             `yaml:"description"`
	Size        int                `yaml:"size"`
	Owns        []flexRelationship `yaml:"owns"`
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

type yamlDataAsset struct {
	Name        string             `yaml:"name"`
	Type        string             `yaml:"type"`
	Description string             `yaml:"description"`
	UsedBy      []yamlDataAssetRef `yaml:"usedBy"`
	ProducedBy  string             `yaml:"producedBy"`
	ConsumedBy  []string           `yaml:"consumedBy"`
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

	if err := addActors(model, raw.Actors); err != nil {
		return nil, err
	}
	// raw.Scenarios is silently ignored (deprecated in 1.9.1).
	if err := addNeeds(model, raw.Needs); err != nil {
		return nil, err
	}
	if err := addCapabilities(model, raw.Capabilities); err != nil {
		return nil, err
	}
	if err := addServices(model, raw.Services); err != nil {
		return nil, err
	}
	if err := addTeams(model, raw.Teams); err != nil {
		return nil, err
	}
	if err := addPlatforms(model, raw.Platforms); err != nil {
		return nil, err
	}
	if err := addInteractions(model, raw.Interactions); err != nil {
		return nil, err
	}
	if err := addDataAssets(model, raw.DataAssets); err != nil {
		return nil, err
	}
	if err := addExternalDependencies(model, raw.ExternalDependencies); err != nil {
		return nil, err
	}

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
		if n.Actor == "" {
			return fmt.Errorf("parser: need %q: need.actor is required", n.Name)
		}
		// n.Scenario is silently ignored (deprecated in 1.9.1).
		need, err := entity.NewNeed(n.Name, n.Name, n.Actor, n.Outcome)
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

func addCapabilities(model *entity.UNMModel, caps []yamlCapability) error {
	for _, c := range caps {
		cap, err := buildCapability(c)
		if err != nil {
			return err
		}
		if err := model.AddCapability(cap); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
}

func buildCapability(yc yamlCapability) (*entity.Capability, error) {
	if yc.Name == "" {
		return nil, fmt.Errorf("parser: capability.name is required")
	}
	cap, err := entity.NewCapability(yc.Name, yc.Name, yc.Description)
	if err != nil {
		return nil, fmt.Errorf("parser: capability %q: %w", yc.Name, err)
	}

	if yc.Visibility != "" {
		if err := cap.SetVisibility(yc.Visibility); err != nil {
			return nil, fmt.Errorf("parser: capability %q: %w", yc.Name, err)
		}
	}

	// Build children first (bottom-up)
	for _, child := range yc.Children {
		childCap, err := buildCapability(child)
		if err != nil {
			return nil, err
		}
		cap.AddChild(childCap)
	}

	for _, rel := range yc.RealizedBy {
		r, err := buildRelationship(rel)
		if err != nil {
			return nil, fmt.Errorf("parser: capability %q realizedBy: %w", yc.Name, err)
		}
		cap.AddRealizedBy(r)
	}

	for _, rel := range yc.DependsOn {
		r, err := buildRelationship(rel)
		if err != nil {
			return nil, fmt.Errorf("parser: capability %q dependsOn: %w", yc.Name, err)
		}
		cap.AddDependsOn(r)
	}

	return cap, nil
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
		// TODO: emit deprecation warnings in a future version.
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

func addInteractions(model *entity.UNMModel, interactions []yamlInteraction) error {
	for i, inter := range interactions {
		mode, err := valueobject.NewInteractionMode(inter.Mode)
		if err != nil {
			return fmt.Errorf("parser: interaction[%d] mode: %w", i, err)
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

func addDataAssets(model *entity.UNMModel, assets []yamlDataAsset) error {
	for _, a := range assets {
		if a.Name == "" {
			return fmt.Errorf("parser: data_asset.name is required")
		}
		da, err := entity.NewDataAsset(a.Name, a.Name, a.Type, a.Description)
		if err != nil {
			return fmt.Errorf("parser: data_asset %q: %w", a.Name, err)
		}
		for _, u := range a.UsedBy {
			da.AddUsedBy(u.Target, u.Access)
		}
		da.ProducedBy = a.ProducedBy
		for _, c := range a.ConsumedBy {
			da.ConsumedBy = append(da.ConsumedBy, c)
		}
		if err := model.AddDataAsset(da); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
}

func addExternalDependencies(model *entity.UNMModel, deps []yamlExternalDependency) error {
	for _, d := range deps {
		if d.Name == "" {
			return fmt.Errorf("parser: external_dependency.name is required")
		}
		ext, err := entity.NewExternalDependency(d.Name, d.Name, d.Description)
		if err != nil {
			return fmt.Errorf("parser: external_dependency %q: %w", d.Name, err)
		}
		for _, u := range d.UsedBy {
			ext.AddUsedBy(u.Target, u.Description)
		}
		if err := model.AddExternalDependency(ext); err != nil {
			return fmt.Errorf("parser: %w", err)
		}
	}
	return nil
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

package serializer

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// yamlRelationship is the long-form relationship output.
type yamlRelationship struct {
	Target      string `yaml:"target"`
	Description string `yaml:"description,omitempty"`
	Role        string `yaml:"role,omitempty"`
}

type yamlSystem struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description,omitempty"`
	Version      string `yaml:"version,omitempty"`
	LastModified string `yaml:"lastModified,omitempty"`
	Author       string `yaml:"author,omitempty"`
}

type yamlActor struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type yamlNeed struct {
	Name        string `yaml:"name"`
	Actor       any    `yaml:"actor"` // string for single actor, []string for multi-actor
	Outcome     string `yaml:"outcome,omitempty"`
	SupportedBy []any  `yaml:"supportedBy,omitempty"`
}

type yamlCapability struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Visibility  string `yaml:"visibility,omitempty"`
	Parent      string `yaml:"parent,omitempty"`
	DependsOn   []any  `yaml:"dependsOn,omitempty"`
}

type yamlService struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description,omitempty"`
	OwnedBy      string   `yaml:"ownedBy"`
	DependsOn    []any    `yaml:"dependsOn,omitempty"`
	Realizes     []any    `yaml:"realizes,omitempty"`
	ExternalDeps []string `yaml:"externalDeps,omitempty"`
}

type yamlTeamInteract struct {
	With        string `yaml:"with"`
	Mode        string `yaml:"mode"`
	Via         string `yaml:"via,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type yamlTeam struct {
	Name        string             `yaml:"name"`
	Type        string             `yaml:"type"`
	Description string             `yaml:"description,omitempty"`
	Size        int                `yaml:"size,omitempty"`
	Owns        []string           `yaml:"owns,omitempty"`
	Interacts   []yamlTeamInteract `yaml:"interacts,omitempty"`
}

type yamlPlatform struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Teams       []string `yaml:"teams,omitempty"`
}

type yamlDataAsset struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type,omitempty"`
	Description string   `yaml:"description,omitempty"`
	UsedBy      []string `yaml:"usedBy,omitempty"`
}

type yamlExternalDependency struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	UsedBy      []string `yaml:"usedBy,omitempty"`
}

type yamlDocument struct {
	System               yamlSystem              `yaml:"system"`
	Actors               []yamlActor             `yaml:"actors,omitempty"`
	Needs                []yamlNeed              `yaml:"needs,omitempty"`
	Capabilities         []yamlCapability        `yaml:"capabilities,omitempty"`
	Services             []yamlService           `yaml:"services,omitempty"`
	Teams                []yamlTeam              `yaml:"teams,omitempty"`
	Platforms            []yamlPlatform          `yaml:"platforms,omitempty"`
	DataAssets           []yamlDataAsset         `yaml:"data_assets,omitempty"`
	ExternalDependencies []yamlExternalDependency `yaml:"external_dependencies,omitempty"`
}

// MarshalYAML converts a UNMModel to valid YAML that round-trips through the parser.
func MarshalYAML(m *entity.UNMModel) ([]byte, error) {
	doc := yamlDocument{
		System: yamlSystem{
			Name:         m.System.Name,
			Description:  m.System.Description,
			Version:      m.Meta.Version,
			LastModified: m.Meta.LastModified,
			Author:       m.Meta.Author,
		},
	}

	doc.Actors = serializeActors(m)
	doc.Needs = serializeNeeds(m)
	doc.Capabilities = serializeCapabilities(m)
	doc.Services = serializeServices(m)
	doc.Teams = serializeTeams(m)
	doc.Platforms = serializePlatforms(m)
	doc.DataAssets = serializeDataAssets(m)
	doc.ExternalDependencies = serializeExternalDependencies(m)

	return yaml.Marshal(&doc)
}

func serializeActors(m *entity.UNMModel) []yamlActor {
	actors := make([]yamlActor, 0, len(m.Actors))
	for _, a := range m.Actors {
		actors = append(actors, yamlActor{Name: a.Name, Description: a.Description})
	}
	sort.Slice(actors, func(i, j int) bool { return actors[i].Name < actors[j].Name })
	return actors
}

func serializeNeeds(m *entity.UNMModel) []yamlNeed {
	needs := make([]yamlNeed, 0, len(m.Needs))
	for _, n := range m.Needs {
		yn := yamlNeed{
			Name:    n.Name,
			Outcome: n.Outcome,
		}
		if len(n.ActorNames) == 1 {
			yn.Actor = n.ActorNames[0] // scalar string — backward compat
		} else if len(n.ActorNames) > 1 {
			yn.Actor = n.ActorNames // sequence — parser handles both
		}
		yn.SupportedBy = serializeRelationships(n.SupportedBy)
		needs = append(needs, yn)
	}
	sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })
	return needs
}

func serializeCapabilities(m *entity.UNMModel) []yamlCapability {
	// Build parent map: child name → parent name
	parentMap := map[string]string{}
	for _, cap := range m.Capabilities {
		for _, child := range cap.Children {
			parentMap[child.Name] = cap.Name
		}
	}

	// Collect all capabilities (roots + their children recursively)
	var collect func(c *entity.Capability, out *[]yamlCapability)
	collect = func(c *entity.Capability, out *[]yamlCapability) {
		yc := yamlCapability{
			Name:        c.Name,
			Description: c.Description,
			Visibility:  c.Visibility,
			Parent:      parentMap[c.Name],
		}
		yc.DependsOn = serializeRelationships(c.DependsOn)
		*out = append(*out, yc)
		for _, child := range c.Children {
			collect(child, out)
		}
	}

	roots := m.GetRootCapabilities()
	var caps []yamlCapability
	for _, rc := range roots {
		collect(rc, &caps)
	}
	sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
	return caps
}

func serializeServices(m *entity.UNMModel) []yamlService {
	// Build reverse realizes map: service name → list of cap names it realizes
	realizesBySvc := map[string][]any{}
	for _, cap := range m.Capabilities {
		for _, rel := range cap.RealizedBy {
			svcName := rel.TargetID.String()
			if rel.Description == "" && rel.Role == "" {
				realizesBySvc[svcName] = append(realizesBySvc[svcName], cap.Name)
			} else {
				realizesBySvc[svcName] = append(realizesBySvc[svcName], yamlRelationship{
					Target:      cap.Name,
					Description: rel.Description,
					Role:        string(rel.Role),
				})
			}
		}
	}

	// Build externalDeps map: service name → list of ext dep names
	extDepsBySvc := map[string][]string{}
	for _, ed := range m.ExternalDependencies {
		for _, u := range ed.UsedBy {
			extDepsBySvc[u.ServiceName] = append(extDepsBySvc[u.ServiceName], ed.Name)
		}
	}

	services := make([]yamlService, 0, len(m.Services))
	for _, s := range m.Services {
		ys := yamlService{
			Name:        s.Name,
			Description: s.Description,
			OwnedBy:     s.OwnerTeamName,
		}
		ys.DependsOn = serializeRelationships(s.DependsOn)
		if realizes := realizesBySvc[s.Name]; len(realizes) > 0 {
			sort.Slice(realizes, func(i, j int) bool {
				ti, tj := fmt.Sprintf("%v", realizes[i]), fmt.Sprintf("%v", realizes[j])
				return ti < tj
			})
			ys.Realizes = realizes
		}
		if extDeps := extDepsBySvc[s.Name]; len(extDeps) > 0 {
			sort.Strings(extDeps)
			ys.ExternalDeps = extDeps
		}
		services = append(services, ys)
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services
}

func serializeTeams(m *entity.UNMModel) []yamlTeam {
	// Build interactions by from-team
	interactsByTeam := map[string][]yamlTeamInteract{}
	for _, ix := range m.Interactions {
		interactsByTeam[ix.FromTeamName] = append(interactsByTeam[ix.FromTeamName], yamlTeamInteract{
			With:        ix.ToTeamName,
			Mode:        string(ix.Mode),
			Via:         ix.Via,
			Description: ix.Description,
		})
	}

	teams := make([]yamlTeam, 0, len(m.Teams))
	for _, t := range m.Teams {
		yt := yamlTeam{
			Name:        t.Name,
			Type:        string(t.TeamType),
			Description: t.Description,
		}
		if t.SizeExplicit {
			yt.Size = t.Size
		}
		for _, rel := range t.Owns {
			yt.Owns = append(yt.Owns, rel.TargetID.String())
		}
		sort.Strings(yt.Owns)
		if interacts := interactsByTeam[t.Name]; len(interacts) > 0 {
			sort.Slice(interacts, func(i, j int) bool { return interacts[i].With < interacts[j].With })
			yt.Interacts = interacts
		}
		teams = append(teams, yt)
	}
	sort.Slice(teams, func(i, j int) bool { return teams[i].Name < teams[j].Name })
	return teams
}

func serializePlatforms(m *entity.UNMModel) []yamlPlatform {
	platforms := make([]yamlPlatform, 0, len(m.Platforms))
	for _, p := range m.Platforms {
		yp := yamlPlatform{
			Name:        p.Name,
			Description: p.Description,
		}
		yp.Teams = make([]string, len(p.TeamNames))
		copy(yp.Teams, p.TeamNames)
		sort.Strings(yp.Teams)
		platforms = append(platforms, yp)
	}
	sort.Slice(platforms, func(i, j int) bool { return platforms[i].Name < platforms[j].Name })
	return platforms
}

func serializeDataAssets(m *entity.UNMModel) []yamlDataAsset {
	assets := make([]yamlDataAsset, 0, len(m.DataAssets))
	for _, da := range m.DataAssets {
		ya := yamlDataAsset{
			Name:        da.Name,
			Type:        da.Type,
			Description: da.Description,
		}
		if len(da.UsedBy) > 0 {
			ya.UsedBy = make([]string, len(da.UsedBy))
			copy(ya.UsedBy, da.UsedBy)
		}
		assets = append(assets, ya)
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Name < assets[j].Name })
	return assets
}

func serializeExternalDependencies(m *entity.UNMModel) []yamlExternalDependency {
	deps := make([]yamlExternalDependency, 0, len(m.ExternalDependencies))
	for _, ed := range m.ExternalDependencies {
		yd := yamlExternalDependency{
			Name:        ed.Name,
			Description: ed.Description,
			// v2: usedBy is emitted on services as externalDeps, not here
		}
		deps = append(deps, yd)
	}
	sort.Slice(deps, func(i, j int) bool { return deps[i].Name < deps[j].Name })
	return deps
}

// serializeRelationships converts entity relationships to YAML-compatible format.
// Uses short form (plain string) when only the target is set,
// long form (object) when description or role is present.
func serializeRelationships(rels []entity.Relationship) []any {
	if len(rels) == 0 {
		return nil
	}
	result := make([]any, 0, len(rels))
	for _, rel := range rels {
		target := rel.TargetID.String()
		if rel.Description == "" && rel.Role == "" {
			result = append(result, target)
		} else {
			result = append(result, yamlRelationship{
				Target:      target,
				Description: rel.Description,
				Role:        string(rel.Role),
			})
		}
	}
	return result
}

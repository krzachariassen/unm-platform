package serializer

import (
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/uber/unm-platform/internal/domain/entity"
)

// yamlRelationship is the long-form relationship output.
type yamlRelationship struct {
	Target      string `yaml:"target"`
	Description string `yaml:"description,omitempty"`
	Role        string `yaml:"role,omitempty"`
}

type yamlSystem struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type yamlActor struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type yamlNeed struct {
	Name        string `yaml:"name"`
	Actor       string `yaml:"actor"`
	Outcome     string `yaml:"outcome,omitempty"`
	SupportedBy []any  `yaml:"supportedBy,omitempty"`
}

type yamlCapability struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description,omitempty"`
	Visibility  string           `yaml:"visibility,omitempty"`
	RealizedBy  []any            `yaml:"realizedBy,omitempty"`
	DependsOn   []any            `yaml:"dependsOn,omitempty"`
	Children    []yamlCapability `yaml:"children,omitempty"`
}

type yamlService struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	OwnedBy     string `yaml:"ownedBy"`
	DependsOn   []any  `yaml:"dependsOn,omitempty"`
}

type yamlTeam struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description,omitempty"`
	Size        int      `yaml:"size,omitempty"`
	Owns        []string `yaml:"owns,omitempty"`
}

type yamlPlatform struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Teams       []string `yaml:"teams,omitempty"`
}

type yamlInteraction struct {
	From        string `yaml:"from"`
	To          string `yaml:"to"`
	Mode        string `yaml:"mode"`
	Via         string `yaml:"via,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type yamlDataAsset struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type,omitempty"`
	Description string   `yaml:"description,omitempty"`
	ProducedBy  string   `yaml:"producedBy,omitempty"`
	ConsumedBy  []string `yaml:"consumedBy,omitempty"`
	UsedBy      []string `yaml:"usedBy,omitempty"`
}

type yamlExternalDependency struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	UsedBy      []string `yaml:"usedBy,omitempty"`
}

type yamlDocument struct {
	System               yamlSystem               `yaml:"system"`
	Actors               []yamlActor              `yaml:"actors,omitempty"`
	Needs                []yamlNeed               `yaml:"needs,omitempty"`
	Capabilities         []yamlCapability          `yaml:"capabilities,omitempty"`
	Services             []yamlService             `yaml:"services,omitempty"`
	Teams                []yamlTeam                `yaml:"teams,omitempty"`
	Platforms            []yamlPlatform            `yaml:"platforms,omitempty"`
	Interactions         []yamlInteraction         `yaml:"interactions,omitempty"`
	DataAssets           []yamlDataAsset           `yaml:"data_assets,omitempty"`
	ExternalDependencies []yamlExternalDependency  `yaml:"external_dependencies,omitempty"`
}

// MarshalYAML converts a UNMModel to valid YAML that round-trips through the parser.
func MarshalYAML(m *entity.UNMModel) ([]byte, error) {
	doc := yamlDocument{
		System: yamlSystem{
			Name:        m.System.Name,
			Description: m.System.Description,
		},
	}

	doc.Actors = serializeActors(m)
	doc.Needs = serializeNeeds(m)
	doc.Capabilities = serializeCapabilities(m)
	doc.Services = serializeServices(m)
	doc.Teams = serializeTeams(m)
	doc.Platforms = serializePlatforms(m)
	doc.Interactions = serializeInteractions(m)
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
			Actor:   n.ActorName,
			Outcome: n.Outcome,
		}
		yn.SupportedBy = serializeRelationships(n.SupportedBy)
		needs = append(needs, yn)
	}
	sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })
	return needs
}

func serializeCapabilities(m *entity.UNMModel) []yamlCapability {
	roots := m.GetRootCapabilities()
	caps := make([]yamlCapability, 0, len(roots))
	for _, rc := range roots {
		caps = append(caps, serializeCapability(rc))
	}
	sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
	return caps
}

func serializeCapability(c *entity.Capability) yamlCapability {
	yc := yamlCapability{
		Name:        c.Name,
		Description: c.Description,
		Visibility:  c.Visibility,
	}
	yc.RealizedBy = serializeRelationships(c.RealizedBy)
	yc.DependsOn = serializeRelationships(c.DependsOn)
	if len(c.Children) > 0 {
		yc.Children = make([]yamlCapability, 0, len(c.Children))
		for _, child := range c.Children {
			yc.Children = append(yc.Children, serializeCapability(child))
		}
		sort.Slice(yc.Children, func(i, j int) bool { return yc.Children[i].Name < yc.Children[j].Name })
	}
	return yc
}

func serializeServices(m *entity.UNMModel) []yamlService {
	services := make([]yamlService, 0, len(m.Services))
	for _, s := range m.Services {
		ys := yamlService{
			Name:        s.Name,
			Description: s.Description,
			OwnedBy:     s.OwnerTeamName,
		}
		ys.DependsOn = serializeRelationships(s.DependsOn)
		services = append(services, ys)
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services
}

func serializeTeams(m *entity.UNMModel) []yamlTeam {
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

func serializeInteractions(m *entity.UNMModel) []yamlInteraction {
	interactions := make([]yamlInteraction, 0, len(m.Interactions))
	for _, ix := range m.Interactions {
		interactions = append(interactions, yamlInteraction{
			From:        ix.FromTeamName,
			To:          ix.ToTeamName,
			Mode:        string(ix.Mode),
			Via:         ix.Via,
			Description: ix.Description,
		})
	}
	sort.Slice(interactions, func(i, j int) bool {
		if interactions[i].From != interactions[j].From {
			return interactions[i].From < interactions[j].From
		}
		return interactions[i].To < interactions[j].To
	})
	return interactions
}

func serializeDataAssets(m *entity.UNMModel) []yamlDataAsset {
	assets := make([]yamlDataAsset, 0, len(m.DataAssets))
	for _, da := range m.DataAssets {
		ya := yamlDataAsset{
			Name:        da.Name,
			Type:        da.Type,
			Description: da.Description,
			ProducedBy:  da.ProducedBy,
		}
		if len(da.ConsumedBy) > 0 {
			ya.ConsumedBy = make([]string, len(da.ConsumedBy))
			copy(ya.ConsumedBy, da.ConsumedBy)
		}
		for _, u := range da.UsedBy {
			ya.UsedBy = append(ya.UsedBy, u.ServiceName)
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
		}
		for _, u := range ed.UsedBy {
			yd.UsedBy = append(yd.UsedBy, u.ServiceName)
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

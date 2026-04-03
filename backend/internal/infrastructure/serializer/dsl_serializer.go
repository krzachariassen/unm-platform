package serializer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// MarshalDSL converts a UNMModel to valid .unm DSL text that round-trips through the parser.
func MarshalDSL(m *entity.UNMModel) ([]byte, error) {
	var b strings.Builder

	writeSystem(&b, m)
	writeActors(&b, m)
	writeNeeds(&b, m)
	writeCapabilities(&b, m)
	writeServicesDSL(&b, m)
	writeTeamsDSL(&b, m)
	writePlatformsDSL(&b, m)
	writeDataAssetsDSL(&b, m)
	writeExternalDependenciesDSL(&b, m)
	writeSignals(&b, m)
	writeInferredMappings(&b, m)
	writeTransitions(&b, m)

	return []byte(b.String()), nil
}

// q wraps a string in double quotes for DSL output.
// The DSL parser does not support escape sequences, so embedded double quotes
// are replaced with single quotes to avoid breaking the grammar.
func q(s string) string {
	safe := strings.ReplaceAll(s, `"`, `'`)
	return `"` + safe + `"`
}

func writeSystem(b *strings.Builder, m *entity.UNMModel) {
	b.WriteString("system ")
	b.WriteString(q(m.System.Name))
	b.WriteString(" {\n")
	if m.System.Description != "" {
		b.WriteString("  description ")
		b.WriteString(q(m.System.Description))
		b.WriteString("\n")
	}
	if m.Meta.Version != "" {
		b.WriteString("  version ")
		b.WriteString(q(m.Meta.Version))
		b.WriteString("\n")
	}
	if m.Meta.LastModified != "" {
		b.WriteString("  lastModified ")
		b.WriteString(q(m.Meta.LastModified))
		b.WriteString("\n")
	}
	if m.Meta.Author != "" {
		b.WriteString("  author ")
		b.WriteString(q(m.Meta.Author))
		b.WriteString("\n")
	}
	b.WriteString("}\n")
}

func writeActors(b *strings.Builder, m *entity.UNMModel) {
	actors := sortedKeys(m.Actors)
	if len(actors) == 0 {
		return
	}
	b.WriteString("\n# ── ACTORS ──\n")
	for _, name := range actors {
		a := m.Actors[name]
		b.WriteString("\nactor ")
		b.WriteString(q(a.Name))
		b.WriteString(" {\n")
		if a.Description != "" {
			b.WriteString("  description ")
			b.WriteString(q(a.Description))
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}
}

func writeNeeds(b *strings.Builder, m *entity.UNMModel) {
	needs := sortedKeys(m.Needs)
	if len(needs) == 0 {
		return
	}
	b.WriteString("\n# ── NEEDS ──\n")
	for _, name := range needs {
		n := m.Needs[name]
		b.WriteString("\nneed ")
		b.WriteString(q(n.Name))
		b.WriteString(" {\n")
		if len(n.ActorNames) == 1 {
			b.WriteString("  actor ")
			b.WriteString(q(n.ActorNames[0]))
			b.WriteString("\n")
		} else if len(n.ActorNames) > 1 {
			b.WriteString("  actor ")
			for i, a := range n.ActorNames {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(q(a))
			}
			b.WriteString("\n")
		}
		if n.Outcome != "" {
			b.WriteString("  outcome ")
			b.WriteString(q(n.Outcome))
			b.WriteString("\n")
		}
		for _, rel := range n.SupportedBy {
			b.WriteString("  supportedBy ")
			writeRelationship(b, rel)
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}
}

func writeCapabilities(b *strings.Builder, m *entity.UNMModel) {
	roots := m.GetRootCapabilities()
	if len(roots) == 0 {
		return
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].Name < roots[j].Name })

	layerOrder := []string{"user-facing", "domain", "foundational", "infrastructure", ""}
	layerLabel := map[string]string{
		"user-facing":    "USER-FACING",
		"domain":         "DOMAIN",
		"foundational":   "FOUNDATIONAL",
		"infrastructure": "INFRASTRUCTURE",
		"":               "UNCATEGORIZED",
	}

	byLayer := map[string][]*entity.Capability{}
	for _, cap := range roots {
		byLayer[cap.Visibility] = append(byLayer[cap.Visibility], cap)
	}

	for _, layer := range layerOrder {
		caps := byLayer[layer]
		if len(caps) == 0 {
			continue
		}
		sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
		b.WriteString(fmt.Sprintf("\n# ── CAPABILITIES: %s ──\n", layerLabel[layer]))
		for _, cap := range caps {
			writeCapabilityBlock(b, m, cap, 0)
		}
		// Also emit flat-parent children that belong to this layer
		for _, childName := range sortedMapKeys(m.CapabilityParents) {
			child := m.Capabilities[childName]
			if child == nil || child.Visibility != layer {
				continue
			}
			// Only emit children that use the flat parent field (not nested)
			// Nested children are emitted inside their parent's block
			if isNestedChild(m, child) {
				continue
			}
			writeCapabilityBlock(b, m, child, 0)
		}
	}
}

// isNestedChild returns true if the child is part of its parent's Children slice
// (as opposed to using the flat parent field).
func isNestedChild(m *entity.UNMModel, child *entity.Capability) bool {
	parentName, ok := m.CapabilityParents[child.Name]
	if !ok {
		return false
	}
	parent, ok := m.Capabilities[parentName]
	if !ok {
		return false
	}
	for _, c := range parent.Children {
		if c.Name == child.Name {
			return true
		}
	}
	return false
}

func writeCapabilityBlock(b *strings.Builder, m *entity.UNMModel, cap *entity.Capability, indent int) {
	pad := strings.Repeat("  ", indent)
	b.WriteString("\n")
	b.WriteString(pad)
	b.WriteString("capability ")
	b.WriteString(q(cap.Name))
	b.WriteString(" {\n")

	innerPad := pad + "  "
	if cap.Visibility != "" {
		b.WriteString(innerPad)
		b.WriteString("visibility ")
		b.WriteString(cap.Visibility)
		b.WriteString("\n")
	}
	if parentName, ok := m.CapabilityParents[cap.Name]; ok && indent == 0 {
		b.WriteString(innerPad)
		b.WriteString("parent ")
		b.WriteString(q(parentName))
		b.WriteString("\n")
	}
	if cap.Description != "" {
		b.WriteString(innerPad)
		b.WriteString("description ")
		b.WriteString(q(cap.Description))
		b.WriteString("\n")
	}
	for _, rel := range cap.DependsOn {
		b.WriteString(innerPad)
		b.WriteString("dependsOn ")
		writeRelationship(b, rel)
		b.WriteString("\n")
	}

	for _, child := range cap.Children {
		writeCapabilityBlock(b, m, child, indent+1)
	}

	b.WriteString(pad)
	b.WriteString("}\n")
}

func writeServicesDSL(b *strings.Builder, m *entity.UNMModel) {
	if len(m.Services) == 0 {
		return
	}

	// Build externalDeps map: service name → list of ext dep names
	extDepsBySvc := map[string][]string{}
	for _, ed := range m.ExternalDependencies {
		for _, u := range ed.UsedBy {
			extDepsBySvc[u.ServiceName] = append(extDepsBySvc[u.ServiceName], ed.Name)
		}
	}

	// Group services by owner team for readability
	byTeam := map[string][]string{}
	var teamOrder []string
	teamSeen := map[string]bool{}
	svcNames := sortedKeys(m.Services)
	for _, name := range svcNames {
		owner := m.Services[name].OwnerTeamName
		if !teamSeen[owner] {
			teamSeen[owner] = true
			teamOrder = append(teamOrder, owner)
		}
		byTeam[owner] = append(byTeam[owner], name)
	}
	sort.Strings(teamOrder)

	b.WriteString("\n# ── SERVICES ──\n")
	for _, team := range teamOrder {
		svcs := byTeam[team]
		sort.Strings(svcs)
		b.WriteString(fmt.Sprintf("\n# ── SERVICES: %s ──\n", team))
		for _, svcName := range svcs {
			svc := m.Services[svcName]
			b.WriteString("\nservice ")
			b.WriteString(q(svc.Name))
			b.WriteString(" {\n")
			if svc.Description != "" {
				b.WriteString("  description ")
				b.WriteString(q(svc.Description))
				b.WriteString("\n")
			}
			b.WriteString("  ownedBy ")
			b.WriteString(q(svc.OwnerTeamName))
			b.WriteString("\n")
			for _, rel := range svc.DependsOn {
				b.WriteString("  dependsOn ")
				b.WriteString(q(rel.TargetID.String()))
				b.WriteString("\n")
			}
			if len(svc.Realizes) > 0 {
				caps := make([]string, 0, len(svc.Realizes))
				for _, rel := range svc.Realizes {
					caps = append(caps, rel.TargetID.String())
				}
				sort.Strings(caps)
				for _, capName := range caps {
					b.WriteString("  realizes ")
					b.WriteString(q(capName))
					b.WriteString("\n")
				}
			}
			if deps := extDepsBySvc[svc.Name]; len(deps) > 0 {
				sort.Strings(deps)
				for _, depName := range deps {
					b.WriteString("  externalDeps ")
					b.WriteString(q(depName))
					b.WriteString("\n")
				}
			}
			b.WriteString("}\n")
		}
	}
}

func writeTeamsDSL(b *strings.Builder, m *entity.UNMModel) {
	if len(m.Teams) == 0 {
		return
	}

	interactsByTeam := map[string][]*entity.Interaction{}
	for _, ix := range m.Interactions {
		interactsByTeam[ix.FromTeamName] = append(interactsByTeam[ix.FromTeamName], ix)
	}

	b.WriteString("\n# ── TEAMS ──\n")
	teamNames := sortedKeys(m.Teams)
	for _, name := range teamNames {
		t := m.Teams[name]
		b.WriteString("\nteam ")
		b.WriteString(q(t.Name))
		b.WriteString(" {\n")
		b.WriteString("  type ")
		b.WriteString(string(t.TeamType))
		b.WriteString("\n")
		if t.SizeExplicit {
			b.WriteString(fmt.Sprintf("  size %d\n", t.Size))
		}
		if t.Description != "" {
			b.WriteString("  description ")
			b.WriteString(q(t.Description))
			b.WriteString("\n")
		}
		var ownNames []string
		for _, rel := range t.Owns {
			ownNames = append(ownNames, rel.TargetID.String())
		}
		sort.Strings(ownNames)
		for _, o := range ownNames {
			b.WriteString("  owns ")
			b.WriteString(q(o))
			b.WriteString("\n")
		}
		if interactions := interactsByTeam[t.Name]; len(interactions) > 0 {
			sort.Slice(interactions, func(i, j int) bool {
				return interactions[i].ToTeamName < interactions[j].ToTeamName
			})
			for _, ix := range interactions {
				b.WriteString("  interacts ")
				b.WriteString(q(ix.ToTeamName))
				b.WriteString(" mode ")
				b.WriteString(string(ix.Mode))
				if ix.Via != "" {
					b.WriteString(" via ")
					b.WriteString(q(ix.Via))
				}
				if ix.Description != "" {
					b.WriteString(" description ")
					b.WriteString(q(ix.Description))
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("}\n")
	}
}

func writePlatformsDSL(b *strings.Builder, m *entity.UNMModel) {
	if len(m.Platforms) == 0 {
		return
	}
	b.WriteString("\n# ── PLATFORMS ──\n")
	platformNames := sortedKeys(m.Platforms)
	for _, name := range platformNames {
		p := m.Platforms[name]
		b.WriteString("\nplatform ")
		b.WriteString(q(p.Name))
		b.WriteString(" {\n")
		if p.Description != "" {
			b.WriteString("  description ")
			b.WriteString(q(p.Description))
			b.WriteString("\n")
		}
		if len(p.TeamNames) > 0 {
			teams := make([]string, len(p.TeamNames))
			copy(teams, p.TeamNames)
			sort.Strings(teams)
			b.WriteString("  teams [")
			for i, t := range teams {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(q(t))
			}
			b.WriteString("]\n")
		}
		b.WriteString("}\n")
	}
}

func writeDataAssetsDSL(b *strings.Builder, m *entity.UNMModel) {
	if len(m.DataAssets) == 0 {
		return
	}
	b.WriteString("\n# ── DATA ASSETS ──\n")
	daNames := sortedKeys(m.DataAssets)
	for _, name := range daNames {
		da := m.DataAssets[name]
		b.WriteString("\ndata_asset ")
		b.WriteString(q(da.Name))
		b.WriteString(" {\n")
		if da.Type != "" {
			b.WriteString("  type ")
			b.WriteString(da.Type)
			b.WriteString("\n")
		}
		if da.Description != "" {
			b.WriteString("  description ")
			b.WriteString(q(da.Description))
			b.WriteString("\n")
		}
		if len(da.UsedBy) > 0 {
			usedBy := make([]string, len(da.UsedBy))
			copy(usedBy, da.UsedBy)
			sort.Strings(usedBy)
			for _, u := range usedBy {
				b.WriteString("  usedBy ")
				b.WriteString(q(u))
				b.WriteString("\n")
			}
		}
		b.WriteString("}\n")
	}
}

func writeExternalDependenciesDSL(b *strings.Builder, m *entity.UNMModel) {
	if len(m.ExternalDependencies) == 0 {
		return
	}
	b.WriteString("\n# ── EXTERNAL DEPENDENCIES ──\n")
	depNames := sortedKeys(m.ExternalDependencies)
	for _, name := range depNames {
		ed := m.ExternalDependencies[name]
		b.WriteString("\nexternal_dependency ")
		b.WriteString(q(ed.Name))
		b.WriteString(" {\n")
		if ed.Description != "" {
			b.WriteString("  description ")
			b.WriteString(q(ed.Description))
			b.WriteString("\n")
		}
		if len(ed.UsedBy) > 0 {
			usages := make([]entity.ExternalUsage, len(ed.UsedBy))
			copy(usages, ed.UsedBy)
			sort.Slice(usages, func(i, j int) bool {
				return usages[i].ServiceName < usages[j].ServiceName
			})
			for _, u := range usages {
				b.WriteString("  usedBy ")
				b.WriteString(q(u.ServiceName))
				if u.Description != "" {
					b.WriteString(" : ")
					b.WriteString(q(u.Description))
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("}\n")
	}
}

func writeSignals(b *strings.Builder, m *entity.UNMModel) {
	if len(m.Signals) == 0 {
		return
	}
	b.WriteString("\n# ── SIGNALS ──\n")
	for _, sig := range m.Signals {
		b.WriteString("\nsignal ")
		b.WriteString(q(sig.ID.String()))
		b.WriteString(" {\n")
		b.WriteString("  category ")
		b.WriteString(q(sig.Category))
		b.WriteString("\n")
		b.WriteString("  severity ")
		b.WriteString(q(string(sig.Severity)))
		b.WriteString("\n")
		b.WriteString("  onEntity ")
		b.WriteString(q(sig.OnEntityName))
		b.WriteString("\n")
		if sig.Description != "" {
			b.WriteString("  description ")
			b.WriteString(q(sig.Description))
			b.WriteString("\n")
		}
		for _, affected := range sig.AffectedEntities {
			b.WriteString("  affects ")
			b.WriteString(q(affected))
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}
}

func writeInferredMappings(b *strings.Builder, m *entity.UNMModel) {
	if len(m.InferredMappings) == 0 {
		return
	}
	b.WriteString("\n# ── INFERRED MAPPINGS ──\n")
	for _, im := range m.InferredMappings {
		b.WriteString("\ninferred {\n")
		b.WriteString("  from ")
		b.WriteString(q(im.ServiceName))
		b.WriteString("\n")
		b.WriteString("  to ")
		b.WriteString(q(im.CapabilityName))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  confidence %.2f\n", im.Confidence.Score))
		if im.Confidence.Evidence != "" {
			b.WriteString("  evidence ")
			b.WriteString(q(im.Confidence.Evidence))
			b.WriteString("\n")
		}
		b.WriteString("  status ")
		b.WriteString(q(string(im.Status)))
		b.WriteString("\n")
		b.WriteString("}\n")
	}
}

func writeTransitions(b *strings.Builder, m *entity.UNMModel) {
	if len(m.Transitions) == 0 {
		return
	}
	b.WriteString("\n# ── TRANSITIONS ──\n")
	for _, tr := range m.Transitions {
		b.WriteString("\ntransition ")
		b.WriteString(q(tr.Name))
		b.WriteString(" {\n")
		if tr.Description != "" {
			b.WriteString("  description ")
			b.WriteString(q(tr.Description))
			b.WriteString("\n")
		}
		if len(tr.Current) > 0 {
			b.WriteString("  current {\n")
			for _, bind := range tr.Current {
				b.WriteString("    capability ")
				b.WriteString(q(bind.CapabilityName))
				b.WriteString(" ownedBy team ")
				b.WriteString(q(bind.TeamName))
				b.WriteString("\n")
			}
			b.WriteString("  }\n")
		}
		if len(tr.Target) > 0 {
			b.WriteString("  target {\n")
			for _, bind := range tr.Target {
				b.WriteString("    capability ")
				b.WriteString(q(bind.CapabilityName))
				b.WriteString(" ownedBy team ")
				b.WriteString(q(bind.TeamName))
				b.WriteString("\n")
			}
			b.WriteString("  }\n")
		}
		for _, step := range tr.Steps {
			b.WriteString(fmt.Sprintf("  step %d ", step.Number))
			b.WriteString(q(step.Label))
			b.WriteString(" {\n")
			if step.ActionText != "" {
				b.WriteString("    action ")
				b.WriteString(step.ActionText)
				b.WriteString("\n")
			}
			if step.ExpectedOutcome != "" {
				b.WriteString("    expected_outcome ")
				b.WriteString(q(step.ExpectedOutcome))
				b.WriteString("\n")
			}
			b.WriteString("  }\n")
		}
		b.WriteString("}\n")
	}
}

func writeRelationship(b *strings.Builder, rel entity.Relationship) {
	b.WriteString(q(rel.TargetID.String()))
	if rel.Description != "" && rel.Role == "" {
		b.WriteString(" : ")
		b.WriteString(q(rel.Description))
	} else if rel.Description != "" || rel.Role != "" {
		b.WriteString(" {\n")
		if rel.Description != "" {
			b.WriteString("    description ")
			b.WriteString(q(rel.Description))
			b.WriteString("\n")
		}
		if rel.Role != "" {
			b.WriteString("    role ")
			b.WriteString(q(string(rel.Role)))
			b.WriteString("\n")
		}
		b.WriteString("  }")
	}
}

// sortedKeys returns the sorted keys of a map[string]*T.
func sortedKeys[T any](m map[string]*T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

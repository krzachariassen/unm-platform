package service

import (
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ModelDiff holds a structured diff between two model versions.
type ModelDiff struct {
	FromVersion int
	ToVersion   int
	Added       DiffEntities
	Removed     DiffEntities
	Changed     DiffEntities
}

// DiffEntities groups entity names by type for a single diff direction.
type DiffEntities struct {
	Actors       []string `json:"actors"`
	Needs        []string `json:"needs"`
	Capabilities []string `json:"capabilities"`
	Services     []string `json:"services"`
	Teams        []string `json:"teams"`
}

// Diff computes a structural diff between two UNM models by comparing entity names.
// Added = names in `to` but not `from`.
// Removed = names in `from` but not `to`.
// Changed = left empty for Phase 14B (deep field diffing is Phase 14C+).
func Diff(from, to *entity.UNMModel) *ModelDiff {
	diff := &ModelDiff{}

	diff.Added.Actors, diff.Removed.Actors = diffNames(mapKeys(actorKeys(from)), mapKeys(actorKeys(to)))
	diff.Added.Needs, diff.Removed.Needs = diffNames(mapKeys(needKeys(from)), mapKeys(needKeys(to)))
	diff.Added.Capabilities, diff.Removed.Capabilities = diffNames(mapKeys(capKeys(from)), mapKeys(capKeys(to)))
	diff.Added.Services, diff.Removed.Services = diffNames(mapKeys(serviceKeys(from)), mapKeys(serviceKeys(to)))
	diff.Added.Teams, diff.Removed.Teams = diffNames(mapKeys(teamKeys(from)), mapKeys(teamKeys(to)))

	return diff
}

// diffNames returns (added, removed) slices given from-names and to-names.
func diffNames(from, to []string) (added, removed []string) {
	fromSet := makeNameSet(from)
	toSet := makeNameSet(to)

	for name := range toSet {
		if !fromSet[name] {
			added = append(added, name)
		}
	}
	for name := range fromSet {
		if !toSet[name] {
			removed = append(removed, name)
		}
	}
	return sortedOrNil(added), sortedOrNil(removed)
}

func makeNameSet(names []string) map[string]bool {
	s := make(map[string]bool, len(names))
	for _, n := range names {
		s[n] = true
	}
	return s
}

func sortedOrNil(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
	return s
}

// mapKeys converts a map to a slice of its keys.
func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func actorKeys(m *entity.UNMModel) map[string]*entity.Actor {
	if m == nil {
		return nil
	}
	return m.Actors
}

func needKeys(m *entity.UNMModel) map[string]*entity.Need {
	if m == nil {
		return nil
	}
	return m.Needs
}

func capKeys(m *entity.UNMModel) map[string]*entity.Capability {
	if m == nil {
		return nil
	}
	return m.Capabilities
}

func serviceKeys(m *entity.UNMModel) map[string]*entity.Service {
	if m == nil {
		return nil
	}
	return m.Services
}

func teamKeys(m *entity.UNMModel) map[string]*entity.Team {
	if m == nil {
		return nil
	}
	return m.Teams
}

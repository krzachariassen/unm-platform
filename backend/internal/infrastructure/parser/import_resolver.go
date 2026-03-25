package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser/dsl"
)

// ParseDSLFile parses a .unm file at the given path and resolves any imports
// relative to the file's directory. Circular imports are detected and rejected.
func ParseDSLFile(path string) (*entity.UNMModel, error) {
	return parseDSLFileWithVisited(path, map[string]bool{})
}

// parseDSLFileWithVisited is the recursive implementation that tracks visited
// absolute paths to detect circular imports.
func parseDSLFileWithVisited(path string, visited map[string]bool) (*entity.UNMModel, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("import resolver: resolve path %q: %w", path, err)
	}
	if visited[abs] {
		return nil, fmt.Errorf("import resolver: circular import detected: %s", path)
	}
	visited[abs] = true

	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("import resolver: read %q: %w", path, err)
	}

	f, err := dsl.Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("import resolver: parse %q: %w", path, err)
	}

	// Transform the current file into a model.
	model, err := dsl.Transform(f)
	if err != nil {
		return nil, fmt.Errorf("import resolver: transform %q: %w", path, err)
	}

	// Resolve imports: parse each imported file and merge its entities.
	baseDir := filepath.Dir(abs)
	for _, imp := range f.Imports {
		importPath := filepath.Join(baseDir, imp.Path)
		importedModel, err := parseDSLFileWithVisited(importPath, visited)
		if err != nil {
			return nil, fmt.Errorf("import resolver: import %q from %q: %w", imp.Path, path, err)
		}
		if err := mergeModel(model, importedModel); err != nil {
			return nil, fmt.Errorf("import resolver: merge import %q: %w", imp.Path, err)
		}
	}

	return model, nil
}

// mergeModel copies all entities from src into dst.
// The dst system name is preserved. Duplicate entity names cause an error.
func mergeModel(dst, src *entity.UNMModel) error {
	for _, a := range src.Actors {
		if err := dst.AddActor(a); err != nil {
			return err
		}
	}
	for _, n := range src.Needs {
		if err := dst.AddNeed(n); err != nil {
			return err
		}
	}
	for name, c := range src.Capabilities {
		// Only add root capabilities; children are added recursively by AddCapability.
		if _, isChild := src.CapabilityParents[name]; isChild {
			continue
		}
		if err := dst.AddCapability(c); err != nil {
			return err
		}
	}
	for _, s := range src.Services {
		if err := dst.AddService(s); err != nil {
			return err
		}
	}
	for _, t := range src.Teams {
		if err := dst.AddTeam(t); err != nil {
			return err
		}
	}
	for _, p := range src.Platforms {
		if err := dst.AddPlatform(p); err != nil {
			return err
		}
	}
	for _, i := range src.Interactions {
		dst.AddInteraction(i)
	}
	for _, s := range src.Signals {
		dst.AddSignal(s)
	}
	for _, d := range src.DataAssets {
		if err := dst.AddDataAsset(d); err != nil {
			return err
		}
	}
	for _, e := range src.ExternalDependencies {
		if err := dst.AddExternalDependency(e); err != nil {
			return err
		}
	}
	for _, im := range src.InferredMappings {
		dst.AddInferredMapping(im)
	}
	for _, tr := range src.Transitions {
		dst.AddTransition(tr)
	}
	return nil
}

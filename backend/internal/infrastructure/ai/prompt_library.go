package ai

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"text/template"
)

//go:embed prompts
var promptFS embed.FS

// PromptLibrary loads and manages prompt templates from embedded files.
type PromptLibrary struct {
	templates map[string]*template.Template
}

// NewPromptLibrary creates a PromptLibrary by loading all .tmpl files from the
// embedded prompts/ directory. Template names are derived from the file path
// without the "prompts/" prefix and ".tmpl" suffix.
func NewPromptLibrary() (*PromptLibrary, error) {
	lib := &PromptLibrary{
		templates: make(map[string]*template.Template),
	}

	err := fs.WalkDir(promptFS, "prompts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		content, err := fs.ReadFile(promptFS, path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		name := strings.TrimPrefix(path, "prompts/")
		name = strings.TrimSuffix(name, ".tmpl")

		tmpl, err := template.New(name).Option("missingkey=zero").Parse(string(content))
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", name, err)
		}

		lib.templates[name] = tmpl
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading prompt templates: %w", err)
	}

	return lib, nil
}

// Get returns a template by name (e.g. "advisor/structural-load").
func (l *PromptLibrary) Get(name string) (*template.Template, error) {
	tmpl, ok := l.templates[name]
	if !ok {
		return nil, fmt.Errorf("template %q not found", name)
	}
	return tmpl, nil
}

// Names returns all available template names, sorted alphabetically.
func (l *PromptLibrary) Names() []string {
	names := make([]string, 0, len(l.templates))
	for name := range l.templates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

package ai

import (
	"bytes"
	"fmt"
)

// PromptRenderer renders named prompt templates with provided data.
type PromptRenderer struct {
	library *PromptLibrary
}

// NewPromptRenderer creates a PromptRenderer using the given PromptLibrary.
func NewPromptRenderer(lib *PromptLibrary) *PromptRenderer {
	return &PromptRenderer{library: lib}
}

// Render executes the named template with data and returns the rendered string.
func (r *PromptRenderer) Render(templateName string, data any) (string, error) {
	tmpl, err := r.library.Get(templateName)
	if err != nil {
		return "", fmt.Errorf("getting template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %q: %w", templateName, err)
	}

	return buf.String(), nil
}

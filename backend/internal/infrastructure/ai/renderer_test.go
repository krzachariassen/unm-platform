package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptRenderer(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)

	renderer := NewPromptRenderer(lib)
	assert.NotNil(t, renderer)
}

func TestPromptRenderer_Render_General(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)

	renderer := NewPromptRenderer(lib)

	// Use map[string]any so template field accesses on missing fields return zero value
	// instead of erroring — templates are designed to handle partial data gracefully.
	data := map[string]any{
		"SystemName": "INCA",
		"Question":   "How?",
	}

	result, err := renderer.Render("advisor/general", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "INCA")
	assert.Contains(t, result, "How?")
}

func TestPromptRenderer_Render_NonexistentTemplate(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)

	renderer := NewPromptRenderer(lib)

	_, err = renderer.Render("nonexistent", nil)
	assert.Error(t, err)
}

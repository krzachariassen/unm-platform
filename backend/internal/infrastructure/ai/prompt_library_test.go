package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptLibrary(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)
	require.NotNil(t, lib)
}

func TestPromptLibrary_Names_ReturnsAllTemplates(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)

	names := lib.Names()
	assert.Len(t, names, 24)

	expected := []string{
		"advisor/bottleneck",
		"advisor/coupling",
		"advisor/fragmentation",
		"advisor/general",
		"advisor/interaction-mode",
		"advisor/need-delivery-risk",
		"advisor/recommendations",
		"advisor/service-placement",
		"advisor/structural-load",
		"advisor/team-boundary",
		"advisor/value-stream",
		"advisor/whatif-scenario",
		"insights/capabilities",
		"insights/cognitive-load",
		"insights/dashboard",
		"insights/needs",
		"insights/ownership",
		"insights/signals",
		"insights/topology",
		"query/health-summary",
		"query/model-summary",
		"query/natural-language",
		"whatif/impact-assessment",
		"whatif/transition-plan",
	}
	assert.Equal(t, expected, names)
}

func TestPromptLibrary_Get_ReturnsTemplate(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)

	tmpl, err := lib.Get("advisor/structural-load")
	require.NoError(t, err)
	assert.NotNil(t, tmpl)
}

func TestPromptLibrary_Get_AllTemplatesExist(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)

	for _, name := range lib.Names() {
		tmpl, err := lib.Get(name)
		assert.NoError(t, err, "template %s should exist", name)
		assert.NotNil(t, tmpl, "template %s should not be nil", name)
	}
}

func TestPromptLibrary_Get_NonexistentReturnsError(t *testing.T) {
	lib, err := NewPromptLibrary()
	require.NoError(t, err)

	tmpl, err := lib.Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, tmpl)
}

package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/unm-platform/internal/domain/entity"
)

func TestTransition_Struct(t *testing.T) {
	tr := &entity.Transition{
		Name:        "Consolidate catalog ownership",
		Description: "Move from fragmented to stream-aligned catalog ownership",
		Current: []entity.TransitionBinding{
			{CapabilityName: "Catalog publication", TeamName: "Team A"},
			{CapabilityName: "Catalog publication", TeamName: "Team B"},
		},
		Target: []entity.TransitionBinding{
			{CapabilityName: "Catalog publication", TeamName: "Catalog Stream"},
		},
		Steps: []entity.TransitionStep{
			{
				Number:          1,
				Label:           "Align Team A and Team B",
				ActionText:      "merge team Team A team Team B into team Catalog Stream",
				ExpectedOutcome: "Single team owns ingestion and validation",
			},
			{
				Number:          2,
				Label:           "Extract platform capabilities",
				ActionText:      "extract capability Catalog storage to team Catalog Platform",
				ExpectedOutcome: "Storage becomes x-as-a-service",
			},
		},
	}

	assert.Equal(t, "Consolidate catalog ownership", tr.Name)
	assert.Equal(t, "Move from fragmented to stream-aligned catalog ownership", tr.Description)
	assert.Len(t, tr.Current, 2)
	assert.Len(t, tr.Target, 1)
	assert.Len(t, tr.Steps, 2)
	assert.Equal(t, 1, tr.Steps[0].Number)
	assert.Equal(t, "Align Team A and Team B", tr.Steps[0].Label)
	assert.Equal(t, 2, tr.Steps[1].Number)
}

func TestAddTransition(t *testing.T) {
	m := entity.NewUNMModel("TestSys", "")

	assert.Empty(t, m.Transitions)

	tr1 := &entity.Transition{Name: "Transition 1"}
	m.AddTransition(tr1)
	assert.Len(t, m.Transitions, 1)
	assert.Equal(t, "Transition 1", m.Transitions[0].Name)

	tr2 := &entity.Transition{Name: "Transition 2"}
	m.AddTransition(tr2)
	assert.Len(t, m.Transitions, 2)
}

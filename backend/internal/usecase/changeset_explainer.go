package usecase

import (
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ChangesetExplainer chains impact analysis and prepares prompt data for AI explanation.
type ChangesetExplainer struct {
	impactAnalyzer ImpactRunner
}

// NewChangesetExplainer constructs a ChangesetExplainer.
func NewChangesetExplainer(ia ImpactRunner) *ChangesetExplainer {
	return &ChangesetExplainer{impactAnalyzer: ia}
}

// PrepareExplainData runs impact analysis on the model+changeset and returns
// a map suitable for rendering the AI explanation prompt template.
func (e *ChangesetExplainer) PrepareExplainData(m *entity.UNMModel, cs *entity.Changeset, changesetID string) (map[string]any, error) {
	impact, err := e.impactAnalyzer.Analyze(m, cs)
	if err != nil {
		return nil, fmt.Errorf("impact analysis failed: %w", err)
	}

	// Format actions as readable text
	actions := make([]string, len(cs.Actions))
	for i, a := range cs.Actions {
		actions[i] = fmt.Sprintf("- %s: %+v", a.Type, a)
	}
	actionsText := ""
	for _, a := range actions {
		actionsText += a + "\n"
	}

	// Format deltas as readable text
	deltasText := ""
	for _, d := range impact.Deltas {
		deltasText += fmt.Sprintf("- %s: %s → %s (%s)\n", d.Dimension, d.Before, d.After, d.Change)
		if d.Detail != "" {
			deltasText += fmt.Sprintf("  Detail: %s\n", d.Detail)
		}
	}

	return map[string]any{
		"SystemName":       m.System.Name,
		"ChangesetID":      changesetID,
		"ChangesetActions": actionsText,
		"ImpactDeltas":     deltasText,
	}, nil
}

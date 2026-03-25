package entity

import (
	"errors"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// Interaction represents a directed relationship between two teams and how they work together.
type Interaction struct {
	ID           valueobject.EntityID
	FromTeamName string
	ToTeamName   string
	Via          string
	Mode         valueobject.InteractionMode
	Description  string
}

// NewInteraction constructs an Interaction. Returns an error if fromTeam or toTeam is empty.
func NewInteraction(id, fromTeam, toTeam string, mode valueobject.InteractionMode, via, description string) (*Interaction, error) {
	if fromTeam == "" {
		return nil, errors.New("interaction: fromTeam must not be empty")
	}
	if toTeam == "" {
		return nil, errors.New("interaction: toTeam must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &Interaction{
		ID:           entityID,
		FromTeamName: fromTeam,
		ToTeamName:   toTeam,
		Via:          via,
		Mode:         mode,
		Description:  description,
	}, nil
}

package entity

import (
	"errors"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// Common data asset type strings. Type is free-form — any string is accepted.
const (
	TypeDatabase    = "database"
	TypeCache       = "cache"
	TypeEventStream = "event-stream"
	TypeBlobStorage = "blob-storage"
	TypeSearchIndex = "search-index"
)

// DataAsset represents a data storage or streaming resource used by services.
// Type is a free-form string. UsedBy is a flat list of service names.
type DataAsset struct {
	ID          valueobject.EntityID
	Name        string
	Type        string
	Description string
	UsedBy      []string
}

// NewDataAsset constructs a DataAsset. Returns an error if name is empty.
// Type is accepted as-is with no validation.
func NewDataAsset(id, name, assetType, description string) (*DataAsset, error) {
	if name == "" {
		return nil, errors.New("data_asset: name must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &DataAsset{
		ID:          entityID,
		Name:        name,
		Type:        assetType,
		Description: description,
		UsedBy:      []string{},
	}, nil
}

// AddUsedBy appends a service name to the UsedBy list.
func (d *DataAsset) AddUsedBy(serviceName string) {
	d.UsedBy = append(d.UsedBy, serviceName)
}


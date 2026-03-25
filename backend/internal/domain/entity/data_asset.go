package entity

import (
	"errors"
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// DataAsset type constants.
const (
	TypeDatabase    = "database"
	TypeCache       = "cache"
	TypeEventStream = "event-stream"
	TypeBlobStorage = "blob-storage"
	TypeSearchIndex = "search-index"
)

var validDataAssetTypes = map[string]bool{
	TypeDatabase:    true,
	TypeCache:       true,
	TypeEventStream: true,
	TypeBlobStorage: true,
	TypeSearchIndex: true,
}

// DataAssetServiceUsage records how a service uses this data asset.
type DataAssetServiceUsage struct {
	ServiceName string
	Access      string
}

// DataAsset represents a data storage or streaming resource used by services.
type DataAsset struct {
	ID          valueobject.EntityID
	Name        string
	Type        string
	Description string
	UsedBy      []DataAssetServiceUsage
	ProducedBy  string
	ConsumedBy  []string
}

// NewDataAsset constructs a DataAsset. Returns an error if name is empty or type is invalid.
func NewDataAsset(id, name, assetType, description string) (*DataAsset, error) {
	if name == "" {
		return nil, errors.New("data_asset: name must not be empty")
	}
	if !validDataAssetTypes[assetType] {
		return nil, fmt.Errorf("data_asset: invalid type %q", assetType)
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
		UsedBy:      []DataAssetServiceUsage{},
		ConsumedBy:  []string{},
	}, nil
}

// AddUsedBy appends a service usage record to UsedBy.
func (d *DataAsset) AddUsedBy(serviceName, access string) {
	d.UsedBy = append(d.UsedBy, DataAssetServiceUsage{
		ServiceName: serviceName,
		Access:      access,
	})
}

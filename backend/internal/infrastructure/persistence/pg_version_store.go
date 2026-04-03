package persistence

import "github.com/jackc/pgx/v5/pgxpool"

// PGVersionStore will implement full version history in Phase 14B.
// For Phase 14A it is a placeholder — version rows are written by PGModelStore.
type PGVersionStore struct {
	db *pgxpool.Pool
}

// NewPGVersionStore creates a PGVersionStore stub.
func NewPGVersionStore(db *pgxpool.Pool) *PGVersionStore {
	return &PGVersionStore{db: db}
}

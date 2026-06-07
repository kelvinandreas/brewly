package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Table represents a physical cafe table with a QR-scannable token.
type Table struct {
	ID           uuid.UUID  `json:"id"`
	Label        string     `json:"label"`
	TokenVersion int        `json:"tokenVersion"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"-"`
}

// TableRepository is the persistence interface for Table.
type TableRepository interface {
	List(ctx context.Context) ([]Table, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Table, error)
	Create(ctx context.Context, t *Table) error
	Update(ctx context.Context, t *Table) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	IncrementTokenVersion(ctx context.Context, id uuid.UUID) error
}

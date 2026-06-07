package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Category groups menu items for display.
type Category struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	DisplayOrder int        `json:"displayOrder"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"-"`
}

// CategoryRepository is the persistence interface for Category.
type CategoryRepository interface {
	List(ctx context.Context) ([]Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Category, error)
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

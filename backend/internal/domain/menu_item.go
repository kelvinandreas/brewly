package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MenuItem represents a single orderable item on the menu.
type MenuItem struct {
	ID          uuid.UUID  `json:"id"`
	CategoryID  uuid.UUID  `json:"categoryId"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	PriceMinor  int64      `json:"priceMinor"`
	ImageURL    *string    `json:"imageUrl"`
	IsAvailable bool       `json:"isAvailable"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"-"`
}

// MenuItemFilter controls which items are returned by List.
type MenuItemFilter struct {
	CategoryID    *uuid.UUID
	AvailableOnly bool
}

// MenuItemRepository is the persistence interface for MenuItem.
type MenuItemRepository interface {
	List(ctx context.Context, filter MenuItemFilter) ([]MenuItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*MenuItem, error)
	Create(ctx context.Context, m *MenuItem) error
	Update(ctx context.Context, m *MenuItem) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

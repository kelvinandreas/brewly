package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Order represents a table's order aggregate.
type Order struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TableID         uuid.UUID   `gorm:"type:uuid;not null"                            json:"tableId"`
	Status          string      `gorm:"not null"                                      json:"status"`
	Source          string      `gorm:"not null"                                      json:"source"`
	TotalMinor      int64       `gorm:"not null"                                      json:"totalMinor"`
	Note            string      `json:"note"`
	CreatedByUserID *uuid.UUID  `gorm:"type:uuid"                                     json:"createdByUserId"`
	Items           []OrderItem `gorm:"foreignKey:OrderID"                            json:"items"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

// OrderItem is a line in an order; snapshots name and price at order time.
type OrderItem struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID            uuid.UUID `gorm:"type:uuid;not null"                            json:"orderId"`
	MenuItemID         uuid.UUID `gorm:"type:uuid;not null"                            json:"menuItemId"`
	NameSnapshot       string    `gorm:"not null"                                      json:"nameSnapshot"`
	PriceMinorSnapshot int64     `gorm:"not null"                                      json:"priceMinorSnapshot"`
	Quantity           int       `gorm:"not null"                                      json:"quantity"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// OrderFilter holds optional query parameters for listing orders.
type OrderFilter struct {
	Status *string
	From   *time.Time
	To     *time.Time
}

// OrderItemInput is the input DTO for creating an order line.
type OrderItemInput struct {
	MenuItemID uuid.UUID
	Quantity   int
}

// OrderRepository defines persistence operations for orders.
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	List(ctx context.Context, filter OrderFilter) ([]Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	ListByTable(ctx context.Context, tableID uuid.UUID, limit int) ([]Order, error)
}

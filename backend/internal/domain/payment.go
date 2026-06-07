package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Payment records a single payment transaction against an order.
type Payment struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID          uuid.UUID `gorm:"type:uuid;not null"                            json:"orderId"`
	Method           string    `gorm:"not null"                                      json:"method"`
	AmountMinor      int64     `gorm:"not null"                                      json:"amountMinor"`
	ReceivedMinor    int64     `gorm:"not null"                                      json:"receivedMinor"`
	RecordedByUserID uuid.UUID `gorm:"type:uuid;not null"                            json:"recordedByUserId"`
	CreatedAt        time.Time `json:"createdAt"`
}

// PaymentRepository defines persistence operations for payments.
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	ListByOrder(ctx context.Context, orderID uuid.UUID) ([]Payment, error)
}

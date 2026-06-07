package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"gorm.io/gorm"
)

type gormPayment struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderID          uuid.UUID `gorm:"type:uuid;not null"`
	Method           string    `gorm:"not null"`
	AmountMinor      int64     `gorm:"not null"`
	ReceivedMinor    int64     `gorm:"not null"`
	RecordedByUserID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt        time.Time
}

func (gormPayment) TableName() string { return "payments" }

func paymentToDomain(g *gormPayment) *domain.Payment {
	return &domain.Payment{
		ID:               g.ID,
		OrderID:          g.OrderID,
		Method:           g.Method,
		AmountMinor:      g.AmountMinor,
		ReceivedMinor:    g.ReceivedMinor,
		RecordedByUserID: g.RecordedByUserID,
		CreatedAt:        g.CreatedAt,
	}
}

// PaymentRepo implements domain.PaymentRepository using GORM.
type PaymentRepo struct {
	db *gorm.DB
}

// NewPaymentRepo constructs a PaymentRepo.
func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

// Create inserts a new payment record.
func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	g := &gormPayment{
		ID:               p.ID,
		OrderID:          p.OrderID,
		Method:           p.Method,
		AmountMinor:      p.AmountMinor,
		ReceivedMinor:    p.ReceivedMinor,
		RecordedByUserID: p.RecordedByUserID,
	}
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("repository.PaymentRepo.Create: %w", err)
	}
	p.CreatedAt = g.CreatedAt
	return nil
}

// ListByOrder returns all payments for a given order.
func (r *PaymentRepo) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]domain.Payment, error) {
	var rows []gormPayment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.PaymentRepo.ListByOrder: %w", err)
	}
	payments := make([]domain.Payment, len(rows))
	for i, g := range rows {
		gCopy := g
		payments[i] = *paymentToDomain(&gCopy)
	}
	return payments, nil
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"gorm.io/gorm"
)

type gormOrder struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	TableID         uuid.UUID `gorm:"type:uuid;not null"`
	Status          string    `gorm:"not null"`
	Source          string    `gorm:"not null"`
	TotalMinor      int64     `gorm:"not null"`
	Note            string
	CreatedByUserID *uuid.UUID      `gorm:"type:uuid"`
	Items           []gormOrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (gormOrder) TableName() string { return "orders" }

type gormOrderItem struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderID            uuid.UUID `gorm:"type:uuid;not null"`
	MenuItemID         uuid.UUID `gorm:"type:uuid;not null"`
	NameSnapshot       string    `gorm:"not null"`
	PriceMinorSnapshot int64     `gorm:"not null"`
	Quantity           int       `gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (gormOrderItem) TableName() string { return "order_items" }

func orderToDomain(g *gormOrder) *domain.Order {
	items := make([]domain.OrderItem, len(g.Items))
	for i, oi := range g.Items {
		items[i] = domain.OrderItem{
			ID:                 oi.ID,
			OrderID:            oi.OrderID,
			MenuItemID:         oi.MenuItemID,
			NameSnapshot:       oi.NameSnapshot,
			PriceMinorSnapshot: oi.PriceMinorSnapshot,
			Quantity:           oi.Quantity,
			CreatedAt:          oi.CreatedAt,
			UpdatedAt:          oi.UpdatedAt,
		}
	}
	return &domain.Order{
		ID:              g.ID,
		TableID:         g.TableID,
		Status:          g.Status,
		Source:          g.Source,
		TotalMinor:      g.TotalMinor,
		Note:            g.Note,
		CreatedByUserID: g.CreatedByUserID,
		Items:           items,
		CreatedAt:       g.CreatedAt,
		UpdatedAt:       g.UpdatedAt,
	}
}

// OrderRepo implements domain.OrderRepository using GORM.
type OrderRepo struct {
	db *gorm.DB
}

// NewOrderRepo constructs an OrderRepo.
func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// Create inserts an order and all its items in a single transaction.
func (r *OrderRepo) Create(ctx context.Context, order *domain.Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	g := &gormOrder{
		ID:              order.ID,
		TableID:         order.TableID,
		Status:          order.Status,
		Source:          order.Source,
		TotalMinor:      order.TotalMinor,
		Note:            order.Note,
		CreatedByUserID: order.CreatedByUserID,
	}
	for i, item := range order.Items {
		if item.ID == uuid.Nil {
			order.Items[i].ID = uuid.New()
		}
		g.Items = append(g.Items, gormOrderItem{
			ID:                 order.Items[i].ID,
			MenuItemID:         item.MenuItemID,
			NameSnapshot:       item.NameSnapshot,
			PriceMinorSnapshot: item.PriceMinorSnapshot,
			Quantity:           item.Quantity,
		})
	}
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("repository.OrderRepo.Create: %w", err)
	}
	order.CreatedAt = g.CreatedAt
	order.UpdatedAt = g.UpdatedAt
	return nil
}

// FindByID returns an order with its items preloaded.
func (r *OrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var g gormOrder
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("repository.OrderRepo.FindByID: %w", err)
	}
	return orderToDomain(&g), nil
}

// List returns orders matching the filter, newest first, with items preloaded.
func (r *OrderRepo) List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, error) {
	q := r.db.WithContext(ctx).Preload("Items")
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.From != nil {
		q = q.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("created_at <= ?", *filter.To)
	}
	var rows []gormOrder
	if err := q.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("repository.OrderRepo.List: %w", err)
	}
	orders := make([]domain.Order, len(rows))
	for i, g := range rows {
		gCopy := g
		orders[i] = *orderToDomain(&gCopy)
	}
	return orders, nil
}

// UpdateStatus sets the status field on a single order.
func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).Model(&gormOrder{}).
		Where("id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("repository.OrderRepo.UpdateStatus: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// ListByTable returns the most recent orders for a table.
func (r *OrderRepo) ListByTable(ctx context.Context, tableID uuid.UUID, limit int) ([]domain.Order, error) {
	var rows []gormOrder
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("table_id = ?", tableID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.OrderRepo.ListByTable: %w", err)
	}
	orders := make([]domain.Order, len(rows))
	for i, g := range rows {
		gCopy := g
		orders[i] = *orderToDomain(&gCopy)
	}
	return orders, nil
}

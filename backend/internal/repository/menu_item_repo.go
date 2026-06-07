package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"gorm.io/gorm"
)

type gormMenuItem struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CategoryID  uuid.UUID  `gorm:"type:uuid"`
	Name        string
	Description *string
	PriceMinor  int64
	ImageURL    *string
	IsAvailable bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`
}

func (gormMenuItem) TableName() string { return "menu_items" }

func menuItemToDomain(g *gormMenuItem) *domain.MenuItem {
	return &domain.MenuItem{
		ID:          g.ID,
		CategoryID:  g.CategoryID,
		Name:        g.Name,
		Description: g.Description,
		PriceMinor:  g.PriceMinor,
		ImageURL:    g.ImageURL,
		IsAvailable: g.IsAvailable,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		DeletedAt:   g.DeletedAt,
	}
}

// MenuItemRepo implements domain.MenuItemRepository using GORM.
type MenuItemRepo struct {
	db *gorm.DB
}

// NewMenuItemRepo constructs a MenuItemRepo.
func NewMenuItemRepo(db *gorm.DB) *MenuItemRepo {
	return &MenuItemRepo{db: db}
}

// List returns active menu items, optionally filtered by category or availability.
func (r *MenuItemRepo) List(ctx context.Context, filter domain.MenuItemFilter) ([]domain.MenuItem, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if filter.CategoryID != nil {
		q = q.Where("category_id = ?", filter.CategoryID)
	}
	if filter.AvailableOnly {
		q = q.Where("is_available = TRUE")
	}
	var rows []gormMenuItem
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("repository.MenuItemRepo.List: %w", err)
	}
	items := make([]domain.MenuItem, len(rows))
	for i, g := range rows {
		items[i] = *menuItemToDomain(&g)
	}
	return items, nil
}

// FindByID returns an active menu item by UUID.
func (r *MenuItemRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	var g gormMenuItem
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMenuItemNotFound
		}
		return nil, fmt.Errorf("repository.MenuItemRepo.FindByID: %w", err)
	}
	return menuItemToDomain(&g), nil
}

// Create inserts a new menu item.
func (r *MenuItemRepo) Create(ctx context.Context, m *domain.MenuItem) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	g := &gormMenuItem{
		ID:          m.ID,
		CategoryID:  m.CategoryID,
		Name:        m.Name,
		Description: m.Description,
		PriceMinor:  m.PriceMinor,
		ImageURL:    m.ImageURL,
		IsAvailable: m.IsAvailable,
	}
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("repository.MenuItemRepo.Create: %w", err)
	}
	m.CreatedAt = g.CreatedAt
	m.UpdatedAt = g.UpdatedAt
	return nil
}

// Update persists mutable fields.
func (r *MenuItemRepo) Update(ctx context.Context, m *domain.MenuItem) error {
	err := r.db.WithContext(ctx).Model(&gormMenuItem{}).
		Where("id = ?", m.ID).
		Updates(map[string]any{
			"category_id":  m.CategoryID,
			"name":         m.Name,
			"description":  m.Description,
			"price_minor":  m.PriceMinor,
			"image_url":    m.ImageURL,
			"is_available": m.IsAvailable,
		}).Error
	if err != nil {
		return fmt.Errorf("repository.MenuItemRepo.Update: %w", err)
	}
	return nil
}

// SoftDelete sets deleted_at to now.
func (r *MenuItemRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Model(&gormMenuItem{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
	if err != nil {
		return fmt.Errorf("repository.MenuItemRepo.SoftDelete: %w", err)
	}
	return nil
}

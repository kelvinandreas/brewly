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

type gormCategory struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name         string
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (gormCategory) TableName() string { return "categories" }

func categoryToDomain(g *gormCategory) *domain.Category {
	return &domain.Category{
		ID:           g.ID,
		Name:         g.Name,
		DisplayOrder: g.DisplayOrder,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		DeletedAt:    g.DeletedAt,
	}
}

// CategoryRepo implements domain.CategoryRepository using GORM.
type CategoryRepo struct {
	db *gorm.DB
}

// NewCategoryRepo constructs a CategoryRepo.
func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

// List returns all active categories ordered by display_order then name.
func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	var rows []gormCategory
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("display_order ASC, name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.CategoryRepo.List: %w", err)
	}
	cats := make([]domain.Category, len(rows))
	for i, g := range rows {
		cats[i] = *categoryToDomain(&g)
	}
	return cats, nil
}

// FindByID returns an active category by UUID.
func (r *CategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var g gormCategory
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("repository.CategoryRepo.FindByID: %w", err)
	}
	return categoryToDomain(&g), nil
}

// Create inserts a new category.
func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	g := &gormCategory{
		ID:           c.ID,
		Name:         c.Name,
		DisplayOrder: c.DisplayOrder,
	}
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("repository.CategoryRepo.Create: %w", err)
	}
	c.CreatedAt = g.CreatedAt
	c.UpdatedAt = g.UpdatedAt
	return nil
}

// Update persists name and display_order changes.
func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	err := r.db.WithContext(ctx).Model(&gormCategory{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"name":          c.Name,
			"display_order": c.DisplayOrder,
		}).Error
	if err != nil {
		return fmt.Errorf("repository.CategoryRepo.Update: %w", err)
	}
	return nil
}

// SoftDelete sets deleted_at to now.
func (r *CategoryRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Model(&gormCategory{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
	if err != nil {
		return fmt.Errorf("repository.CategoryRepo.SoftDelete: %w", err)
	}
	return nil
}

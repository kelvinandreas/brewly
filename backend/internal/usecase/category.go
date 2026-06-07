package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
)

// CategoryUsecase handles business logic for menu categories.
type CategoryUsecase struct {
	repo domain.CategoryRepository
}

// NewCategoryUsecase constructs a CategoryUsecase.
func NewCategoryUsecase(repo domain.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{repo: repo}
}

// List returns all active categories.
func (u *CategoryUsecase) List(ctx context.Context) ([]domain.Category, error) {
	cats, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase.Category.List: %w", err)
	}
	return cats, nil
}

// Create adds a new category.
func (u *CategoryUsecase) Create(ctx context.Context, name string, displayOrder int) (*domain.Category, error) {
	cat := &domain.Category{
		ID:           uuid.New(),
		Name:         name,
		DisplayOrder: displayOrder,
	}
	if err := u.repo.Create(ctx, cat); err != nil {
		return nil, fmt.Errorf("usecase.Category.Create: %w", err)
	}
	return cat, nil
}

// Update patches name and/or display order.
func (u *CategoryUsecase) Update(ctx context.Context, id uuid.UUID, name *string, displayOrder *int) (*domain.Category, error) {
	cat, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.Category.Update find: %w", err)
	}
	if name != nil {
		cat.Name = *name
	}
	if displayOrder != nil {
		cat.DisplayOrder = *displayOrder
	}
	if err := u.repo.Update(ctx, cat); err != nil {
		return nil, fmt.Errorf("usecase.Category.Update: %w", err)
	}
	return cat, nil
}

// SoftDelete removes a category.
func (u *CategoryUsecase) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("usecase.Category.SoftDelete: %w", err)
	}
	return nil
}

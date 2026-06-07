package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
)

// MenuItemUsecase handles business logic for menu items.
type MenuItemUsecase struct {
	repo     domain.MenuItemRepository
	catRepo  domain.CategoryRepository
}

// NewMenuItemUsecase constructs a MenuItemUsecase.
func NewMenuItemUsecase(repo domain.MenuItemRepository, catRepo domain.CategoryRepository) *MenuItemUsecase {
	return &MenuItemUsecase{repo: repo, catRepo: catRepo}
}

// List returns active menu items filtered by category and/or availability.
func (u *MenuItemUsecase) List(ctx context.Context, categoryID *uuid.UUID, availableOnly bool) ([]domain.MenuItem, error) {
	items, err := u.repo.List(ctx, domain.MenuItemFilter{
		CategoryID:    categoryID,
		AvailableOnly: availableOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("usecase.MenuItem.List: %w", err)
	}
	return items, nil
}

// Create adds a new menu item. Validates that the category exists.
func (u *MenuItemUsecase) Create(ctx context.Context, item domain.MenuItem) (*domain.MenuItem, error) {
	if _, err := u.catRepo.FindByID(ctx, item.CategoryID); err != nil {
		return nil, fmt.Errorf("usecase.MenuItem.Create: %w", err)
	}
	item.ID = uuid.New()
	if err := u.repo.Create(ctx, &item); err != nil {
		return nil, fmt.Errorf("usecase.MenuItem.Create: %w", err)
	}
	return &item, nil
}

// Update patches mutable fields of an existing menu item.
func (u *MenuItemUsecase) Update(ctx context.Context, id uuid.UUID, patch domain.MenuItem) (*domain.MenuItem, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.MenuItem.Update find: %w", err)
	}
	if patch.Name != "" {
		item.Name = patch.Name
	}
	if patch.Description != nil {
		item.Description = patch.Description
	}
	if patch.PriceMinor > 0 {
		item.PriceMinor = patch.PriceMinor
	}
	if patch.ImageURL != nil {
		item.ImageURL = patch.ImageURL
	}
	// IsAvailable is a bool — always apply it from the patch (zero value = false is intentional)
	item.IsAvailable = patch.IsAvailable
	if patch.CategoryID != uuid.Nil {
		if _, err := u.catRepo.FindByID(ctx, patch.CategoryID); err != nil {
			return nil, fmt.Errorf("usecase.MenuItem.Update category: %w", err)
		}
		item.CategoryID = patch.CategoryID
	}
	if err := u.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("usecase.MenuItem.Update: %w", err)
	}
	return item, nil
}

// SoftDelete removes a menu item.
func (u *MenuItemUsecase) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("usecase.MenuItem.SoftDelete: %w", err)
	}
	return nil
}

//go:build ignore

// Canonical pattern for a usecase.
// Rules: no HTTP, no GORM. Depends only on domain interfaces.
package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
)

// ThingUsecase holds business logic for Things.
type ThingUsecase struct {
	repo domain.ThingRepository
}

// NewThingUsecase wires dependencies. Called from cmd/api/main.go.
func NewThingUsecase(repo domain.ThingRepository) *ThingUsecase {
	return &ThingUsecase{repo: repo}
}

// CreateThing validates and persists a new Thing.
func (u *ThingUsecase) CreateThing(name string) (*domain.Thing, error) {
	if name == "" {
		return nil, domain.ErrThingConflict
	}

	thing := &domain.Thing{
		ID:     uuid.New(),
		Name:   name,
		Status: domain.ThingStatusActive,
	}

	if err := u.repo.Create(thing); err != nil {
		return nil, fmt.Errorf("ThingUsecase.CreateThing: %w", err)
	}

	return thing, nil
}

// GetThing retrieves a Thing by ID.
func (u *ThingUsecase) GetThing(id uuid.UUID) (*domain.Thing, error) {
	thing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("ThingUsecase.GetThing: %w", err)
	}
	return thing, nil
}

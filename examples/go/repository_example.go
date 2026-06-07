//go:build ignore

// Canonical pattern for a GORM repository.
// Rules: GORM lives here only. Implements a domain interface.
package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"gorm.io/gorm"
)

// thingModel is the GORM model — separate from the domain entity.
type thingModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name      string         `gorm:"not null"`
	Status    string         `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (thingModel) TableName() string { return "things" }

// ThingRepo implements domain.ThingRepository.
type ThingRepo struct {
	db *gorm.DB
}

// NewThingRepo constructs a ThingRepo. Called from cmd/api/main.go.
func NewThingRepo(db *gorm.DB) *ThingRepo {
	return &ThingRepo{db: db}
}

func (r *ThingRepo) Create(thing *domain.Thing) error {
	m := toModel(thing)
	if err := r.db.Create(&m).Error; err != nil {
		return fmt.Errorf("ThingRepo.Create: %w", err)
	}
	return nil
}

func (r *ThingRepo) GetByID(id uuid.UUID) (*domain.Thing, error) {
	var m thingModel
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrThingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ThingRepo.GetByID: %w", err)
	}
	return toEntity(&m), nil
}

func (r *ThingRepo) SoftDelete(id uuid.UUID) error {
	res := r.db.Where("id = ? AND deleted_at IS NULL", id).Delete(&thingModel{})
	if res.Error != nil {
		return fmt.Errorf("ThingRepo.SoftDelete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrThingNotFound
	}
	return nil
}

func toModel(t *domain.Thing) thingModel {
	return thingModel{ID: t.ID, Name: t.Name, Status: string(t.Status)}
}

func toEntity(m *thingModel) *domain.Thing {
	return &domain.Thing{
		ID:        m.ID,
		Name:      m.Name,
		Status:    domain.ThingStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

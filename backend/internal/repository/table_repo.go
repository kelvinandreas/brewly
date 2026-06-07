package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kelvinandreas/brewly/internal/domain"
	"gorm.io/gorm"
)

type gormTable struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Label        string
	TokenVersion int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (gormTable) TableName() string { return "tables" }

func tableToDomain(g *gormTable) *domain.Table {
	return &domain.Table{
		ID:           g.ID,
		Label:        g.Label,
		TokenVersion: g.TokenVersion,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		DeletedAt:    g.DeletedAt,
	}
}

// TableRepo implements domain.TableRepository using GORM.
type TableRepo struct {
	db *gorm.DB
}

// NewTableRepo constructs a TableRepo.
func NewTableRepo(db *gorm.DB) *TableRepo {
	return &TableRepo{db: db}
}

// List returns all active tables ordered by label.
func (r *TableRepo) List(ctx context.Context) ([]domain.Table, error) {
	var rows []gormTable
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("label ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.TableRepo.List: %w", err)
	}
	tables := make([]domain.Table, len(rows))
	for i, g := range rows {
		tables[i] = *tableToDomain(&g)
	}
	return tables, nil
}

// FindByID returns an active table by UUID.
func (r *TableRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Table, error) {
	var g gormTable
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTableNotFound
		}
		return nil, fmt.Errorf("repository.TableRepo.FindByID: %w", err)
	}
	return tableToDomain(&g), nil
}

// Create inserts a new table.
func (r *TableRepo) Create(ctx context.Context, t *domain.Table) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	g := &gormTable{
		ID:           t.ID,
		Label:        t.Label,
		TokenVersion: t.TokenVersion,
	}
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrTableLabelTaken
		}
		return fmt.Errorf("repository.TableRepo.Create: %w", err)
	}
	t.CreatedAt = g.CreatedAt
	t.UpdatedAt = g.UpdatedAt
	return nil
}

// Update persists the label.
func (r *TableRepo) Update(ctx context.Context, t *domain.Table) error {
	err := r.db.WithContext(ctx).Model(&gormTable{}).
		Where("id = ?", t.ID).
		Update("label", t.Label).Error
	if err != nil {
		return fmt.Errorf("repository.TableRepo.Update: %w", err)
	}
	return nil
}

// SoftDelete sets deleted_at to now.
func (r *TableRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Model(&gormTable{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
	if err != nil {
		return fmt.Errorf("repository.TableRepo.SoftDelete: %w", err)
	}
	return nil
}

// IncrementTokenVersion bumps token_version by 1, immediately invalidating all existing table tokens.
func (r *TableRepo) IncrementTokenVersion(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Model(&gormTable{}).
		Where("id = ? AND deleted_at IS NULL", id).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
	if err != nil {
		return fmt.Errorf("repository.TableRepo.IncrementTokenVersion: %w", err)
	}
	return nil
}

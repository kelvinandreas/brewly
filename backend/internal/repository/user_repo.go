// Package repository provides GORM-backed implementations of domain repository interfaces.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/your-handle/brewly/internal/domain"
	"gorm.io/gorm"
)

// gormUser is the GORM model for the users table.
type gormUser struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email          string     `gorm:"uniqueIndex"`
	PasswordHash   string
	Name           string
	Role           string
	LastRefreshJTI *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `gorm:"index"`
}

func (gormUser) TableName() string { return "users" }

func toGorm(u *domain.User) *gormUser {
	return &gormUser{
		ID:             u.ID,
		Email:          u.Email,
		PasswordHash:   u.PasswordHash,
		Name:           u.Name,
		Role:           u.Role,
		LastRefreshJTI: u.LastRefreshJTI,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		DeletedAt:      u.DeletedAt,
	}
}

func toDomain(g *gormUser) *domain.User {
	return &domain.User{
		ID:             g.ID,
		Email:          g.Email,
		PasswordHash:   g.PasswordHash,
		Name:           g.Name,
		Role:           g.Role,
		LastRefreshJTI: g.LastRefreshJTI,
		CreatedAt:      g.CreatedAt,
		UpdatedAt:      g.UpdatedAt,
		DeletedAt:      g.DeletedAt,
	}
}

// UserRepo implements domain.UserRepository using GORM.
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo constructs a UserRepo backed by db.
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user row.
func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	g := toGorm(u)
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("repository.UserRepo.Create: %w", err)
	}
	u.CreatedAt = g.CreatedAt
	u.UpdatedAt = g.UpdatedAt
	return nil
}

// FindByEmail returns an active user with the given email.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var g gormUser
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.UserRepo.FindByEmail: %w", err)
	}
	return toDomain(&g), nil
}

// FindByID returns an active user with the given UUID.
func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var g gormUser
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.UserRepo.FindByID: %w", err)
	}
	return toDomain(&g), nil
}

// ExistsByRole returns true if at least one active user with the given role exists.
func (r *UserRepo) ExistsByRole(ctx context.Context, role string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&gormUser{}).
		Where("role = ? AND deleted_at IS NULL", role).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("repository.UserRepo.ExistsByRole: %w", err)
	}
	return count > 0, nil
}

// SetLastRefreshJTI updates only the last_refresh_jti column for the given user.
func (r *UserRepo) SetLastRefreshJTI(ctx context.Context, id uuid.UUID, jti string) error {
	err := r.db.WithContext(ctx).Model(&gormUser{}).
		Where("id = ?", id).
		Update("last_refresh_jti", jti).Error
	if err != nil {
		return fmt.Errorf("repository.UserRepo.SetLastRefreshJTI: %w", err)
	}
	return nil
}

// Update persists mutable fields of an existing user (name, role, password_hash).
func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	err := r.db.WithContext(ctx).Model(&gormUser{}).
		Where("id = ?", u.ID).
		Updates(map[string]any{
			"name":          u.Name,
			"role":          u.Role,
			"password_hash": u.PasswordHash,
		}).Error
	if err != nil {
		return fmt.Errorf("repository.UserRepo.Update: %w", err)
	}
	return nil
}

// SoftDelete sets deleted_at to now for the given user.
func (r *UserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&gormUser{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
	if err != nil {
		return fmt.Errorf("repository.UserRepo.SoftDelete: %w", err)
	}
	return nil
}

// List returns all active users ordered by created_at.
func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	var rows []gormUser
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.UserRepo.List: %w", err)
	}
	users := make([]domain.User, len(rows))
	for i, g := range rows {
		users[i] = *toDomain(&g)
	}
	return users, nil
}

// isUniqueViolation detects PostgreSQL unique constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

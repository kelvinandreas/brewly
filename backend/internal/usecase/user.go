package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// UserUsecase handles owner-scoped staff account management.
type UserUsecase struct {
	repo domain.UserRepository
}

// NewUserUsecase constructs a UserUsecase.
func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// List returns all active staff accounts.
func (u *UserUsecase) List(ctx context.Context) ([]domain.User, error) {
	users, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase.UserUsecase.List: %w", err)
	}
	return users, nil
}

// Create adds a new cashier or kitchen account.
func (u *UserUsecase) Create(ctx context.Context, email, password, name, role string) (*domain.User, error) {
	if role == domain.RoleOwner {
		return nil, domain.ErrForbidden
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("usecase.UserUsecase.Create hash: %w", err)
	}
	usr := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
	}
	if err := u.repo.Create(ctx, usr); err != nil {
		return nil, fmt.Errorf("usecase.UserUsecase.Create: %w", err)
	}
	return usr, nil
}

// UpdateFields patches name, role, and/or password for an existing staff account.
func (u *UserUsecase) UpdateFields(ctx context.Context, id uuid.UUID, name, role, password *string) (*domain.User, error) {
	usr, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.UserUsecase.UpdateFields find: %w", err)
	}
	if name != nil {
		usr.Name = *name
	}
	if role != nil {
		if *role == domain.RoleOwner {
			return nil, domain.ErrForbidden
		}
		usr.Role = *role
	}
	if password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, fmt.Errorf("usecase.UserUsecase.UpdateFields hash: %w", err)
		}
		usr.PasswordHash = string(hash)
	}
	if err := u.repo.Update(ctx, usr); err != nil {
		return nil, fmt.Errorf("usecase.UserUsecase.UpdateFields update: %w", err)
	}
	return usr, nil
}

// SoftDelete removes a staff account (soft delete).
func (u *UserUsecase) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("usecase.UserUsecase.SoftDelete: %w", err)
	}
	return nil
}

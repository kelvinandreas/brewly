package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User is the staff account entity.
type User struct {
	ID             uuid.UUID  `json:"id"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	Name           string     `json:"name"`
	Role           string     `json:"role"`
	LastRefreshJTI *string    `json:"-"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"-"`
}

// UserRepository is the persistence interface for User. Implemented in repository/.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	ExistsByRole(ctx context.Context, role string) (bool, error)
	SetLastRefreshJTI(ctx context.Context, id uuid.UUID, jti string) error
	Update(ctx context.Context, u *User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]User, error)
}

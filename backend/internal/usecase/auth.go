// Package usecase contains Brewly's business logic layer.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	jwtpkg "github.com/your-handle/brewly/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// AuthConfig holds secrets and TTLs needed by AuthUsecase.
type AuthConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// AuthTokenPair is returned by Login and Refresh.
type AuthTokenPair struct {
	AccessToken  string
	RefreshToken string
}

// AuthUsecase handles staff authentication logic.
type AuthUsecase struct {
	repo domain.UserRepository
	cfg  AuthConfig
}

// NewAuthUsecase constructs an AuthUsecase.
func NewAuthUsecase(repo domain.UserRepository, cfg AuthConfig) *AuthUsecase {
	return &AuthUsecase{repo: repo, cfg: cfg}
}

// RegisterOwner creates the first owner account. Fails if any owner already exists.
func (a *AuthUsecase) RegisterOwner(ctx context.Context, email, password, name string) (*domain.User, error) {
	exists, err := a.repo.ExistsByRole(ctx, domain.RoleOwner)
	if err != nil {
		return nil, fmt.Errorf("usecase.RegisterOwner: %w", err)
	}
	if exists {
		return nil, domain.ErrOwnerExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("usecase.RegisterOwner hash: %w", err)
	}

	u := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         domain.RoleOwner,
	}
	if err := a.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("usecase.RegisterOwner create: %w", err)
	}
	return u, nil
}

// Login verifies credentials and returns a token pair.
func (a *AuthUsecase) Login(ctx context.Context, email, password string) (*domain.User, AuthTokenPair, error) {
	user, err := a.repo.FindByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, AuthTokenPair{}, domain.ErrInvalidCredentials
		}
		return nil, AuthTokenPair{}, fmt.Errorf("usecase.Login find: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, AuthTokenPair{}, domain.ErrInvalidCredentials
	}

	pair, err := a.issuePair(ctx, user)
	if err != nil {
		return nil, AuthTokenPair{}, err
	}
	return user, pair, nil
}

// Refresh validates a refresh token, rotates it, and returns a new pair.
func (a *AuthUsecase) Refresh(ctx context.Context, rawRefreshToken string) (AuthTokenPair, error) {
	claims, err := jwtpkg.Verify(rawRefreshToken, a.cfg.RefreshSecret)
	if err != nil {
		return AuthTokenPair{}, domain.ErrInvalidCredentials
	}

	id, err := uuid.Parse(claims.Sub)
	if err != nil {
		return AuthTokenPair{}, domain.ErrInvalidCredentials
	}

	user, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return AuthTokenPair{}, fmt.Errorf("usecase.Refresh find: %w", err)
	}
	if user.DeletedAt != nil {
		return AuthTokenPair{}, domain.ErrInvalidCredentials
	}

	// Replay / revocation guard — stored JTI must match incoming JTI.
	if user.LastRefreshJTI == nil || *user.LastRefreshJTI != claims.JTI {
		return AuthTokenPair{}, domain.ErrInvalidCredentials
	}

	pair, err := a.issuePair(ctx, user)
	if err != nil {
		return AuthTokenPair{}, err
	}
	return pair, nil
}

// Logout clears the stored refresh JTI, invalidating any outstanding refresh token.
func (a *AuthUsecase) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := a.repo.SetLastRefreshJTI(ctx, userID, ""); err != nil {
		return fmt.Errorf("usecase.Logout: %w", err)
	}
	return nil
}

// Me returns the active user for the given ID.
func (a *AuthUsecase) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := a.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase.Me: %w", err)
	}
	return user, nil
}

// issuePair signs a new access + refresh token and persists the refresh JTI.
func (a *AuthUsecase) issuePair(ctx context.Context, user *domain.User) (AuthTokenPair, error) {
	refreshJTI := uuid.New().String()

	access := jwtpkg.Sign(jwtpkg.Claims{
		Sub:  user.ID.String(),
		Role: user.Role,
		JTI:  uuid.New().String(),
	}, a.cfg.AccessSecret, a.cfg.AccessTTL)

	refresh := jwtpkg.Sign(jwtpkg.Claims{
		Sub: user.ID.String(),
		JTI: refreshJTI,
	}, a.cfg.RefreshSecret, a.cfg.RefreshTTL)

	if err := a.repo.SetLastRefreshJTI(ctx, user.ID, refreshJTI); err != nil {
		return AuthTokenPair{}, fmt.Errorf("usecase.issuePair persist jti: %w", err)
	}
	return AuthTokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

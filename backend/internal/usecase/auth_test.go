package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

// ─── mock repository ──────────────────────────────────────────────────────────

type mockUserRepo struct {
	users          map[uuid.UUID]*domain.User
	emailIndex     map[string]*domain.User
	existsByRole   map[string]bool
	createErr      error
	findByEmailErr error
	findByIDErr    error
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{
		users:        make(map[uuid.UUID]*domain.User),
		emailIndex:   make(map[string]*domain.User),
		existsByRole: make(map[string]bool),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	m.users[u.ID] = u
	m.emailIndex[u.Email] = u
	m.existsByRole[u.Role] = true
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	u, ok := m.emailIndex[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) ExistsByRole(ctx context.Context, role string) (bool, error) {
	return m.existsByRole[role], nil
}

func (m *mockUserRepo) SetLastRefreshJTI(ctx context.Context, id uuid.UUID, jti string) error {
	u, ok := m.users[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.LastRefreshJTI = &jti
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	if _, ok := m.users[u.ID]; !ok {
		return domain.ErrUserNotFound
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	u, ok := m.users[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	now := time.Now()
	u.DeletedAt = &now
	return nil
}

func (m *mockUserRepo) List(ctx context.Context) ([]domain.User, error) {
	out := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, *u)
	}
	return out, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func testCfg() usecase.AuthConfig {
	return usecase.AuthConfig{
		AccessSecret:  "access-secret-32-bytes-long-enough",
		RefreshSecret: "refresh-secret-32-bytes-long-enuf",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
	}
}

func seedOwner(t *testing.T, uc *usecase.AuthUsecase) *domain.User {
	t.Helper()
	u, err := uc.RegisterOwner(context.Background(), "owner@brew.ly", "password123", "Owner")
	if err != nil {
		t.Fatalf("seedOwner: %v", err)
	}
	return u
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestRegisterOwner_success(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())

	u, err := uc.RegisterOwner(context.Background(), "owner@brew.ly", "password123", "Owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Role != domain.RoleOwner {
		t.Errorf("role: want owner, got %s", u.Role)
	}
	if u.PasswordHash == "password123" {
		t.Error("password should be hashed")
	}
}

func TestRegisterOwner_twice_returns_ErrOwnerExists(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	seedOwner(t, uc)

	_, err := uc.RegisterOwner(context.Background(), "another@brew.ly", "password123", "Another")
	if !errors.Is(err, domain.ErrOwnerExists) {
		t.Fatalf("want ErrOwnerExists, got %v", err)
	}
}

func TestLogin_success(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	seedOwner(t, uc)

	user, pair, err := uc.Login(context.Background(), "owner@brew.ly", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "owner@brew.ly" {
		t.Errorf("email mismatch")
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("tokens should not be empty")
	}
}

func TestLogin_wrongPassword_returns_ErrInvalidCredentials(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	seedOwner(t, uc)

	_, _, err := uc.Login(context.Background(), "owner@brew.ly", "wrongpass")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_unknownEmail_returns_ErrInvalidCredentials(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())

	_, _, err := uc.Login(context.Background(), "nobody@brew.ly", "pass")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestRefresh_success(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	seedOwner(t, uc)

	_, pair1, _ := uc.Login(context.Background(), "owner@brew.ly", "password123")

	pair2, err := uc.Refresh(context.Background(), pair1.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair2.AccessToken == "" || pair2.RefreshToken == "" {
		t.Error("rotated tokens should not be empty")
	}
}

func TestRefresh_replayedJTI_returns_ErrInvalidCredentials(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	seedOwner(t, uc)

	_, pair1, _ := uc.Login(context.Background(), "owner@brew.ly", "password123")
	// First refresh rotates the JTI.
	_, _ = uc.Refresh(context.Background(), pair1.RefreshToken)
	// Second refresh with the old token must fail.
	_, err := uc.Refresh(context.Background(), pair1.RefreshToken)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials for replayed token, got %v", err)
	}
}

func TestRefresh_expiredToken_returns_ErrInvalidCredentials(t *testing.T) {
	repo := newMockRepo()
	cfg := testCfg()
	cfg.RefreshTTL = -time.Second // already expired
	uc := usecase.NewAuthUsecase(repo, cfg)
	seedOwner(t, uc)

	_, pair, _ := uc.Login(context.Background(), "owner@brew.ly", "password123")
	_, err := uc.Refresh(context.Background(), pair.RefreshToken)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials for expired token, got %v", err)
	}
}

func TestMe_returnsUser(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	u := seedOwner(t, uc)

	got, err := uc.Me(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("id mismatch")
	}
}

func TestLogout_clearsJTI(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	u := seedOwner(t, uc)

	_, pair, _ := uc.Login(context.Background(), "owner@brew.ly", "password123")

	if err := uc.Logout(context.Background(), u.ID); err != nil {
		t.Fatalf("logout error: %v", err)
	}

	// Refresh after logout must fail (JTI cleared → mismatch).
	_, err := uc.Refresh(context.Background(), pair.RefreshToken)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials after logout, got %v", err)
	}
}

// Ensure bcrypt cost is applied (hash != plaintext).
func TestRegisterOwner_passwordHashed(t *testing.T) {
	repo := newMockRepo()
	uc := usecase.NewAuthUsecase(repo, testCfg())
	u, _ := uc.RegisterOwner(context.Background(), "o@b.ly", "mypassword", "O")
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("mypassword")); err != nil {
		t.Errorf("password hash does not match plaintext: %v", err)
	}
}

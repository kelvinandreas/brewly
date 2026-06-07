package usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/tabletoken"
)

// ─── mock table repository ────────────────────────────────────────────────────

type mockTableRepo struct {
	tables map[uuid.UUID]*domain.Table
}

func newMockTableRepo() *mockTableRepo {
	return &mockTableRepo{tables: make(map[uuid.UUID]*domain.Table)}
}

func (m *mockTableRepo) List(ctx context.Context) ([]domain.Table, error) {
	out := make([]domain.Table, 0)
	for _, t := range m.tables {
		out = append(out, *t)
	}
	return out, nil
}

func (m *mockTableRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Table, error) {
	t, ok := m.tables[id]
	if !ok {
		return nil, domain.ErrTableNotFound
	}
	return t, nil
}

func (m *mockTableRepo) Create(ctx context.Context, t *domain.Table) error {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	m.tables[t.ID] = t
	return nil
}

func (m *mockTableRepo) Update(ctx context.Context, t *domain.Table) error {
	if _, ok := m.tables[t.ID]; !ok {
		return domain.ErrTableNotFound
	}
	m.tables[t.ID] = t
	return nil
}

func (m *mockTableRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	t, ok := m.tables[id]
	if !ok {
		return domain.ErrTableNotFound
	}
	now := time.Now()
	t.DeletedAt = &now
	return nil
}

func (m *mockTableRepo) IncrementTokenVersion(ctx context.Context, id uuid.UUID) error {
	t, ok := m.tables[id]
	if !ok {
		return domain.ErrTableNotFound
	}
	t.TokenVersion++
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func testTableCfg() usecase.TableConfig {
	return usecase.TableConfig{
		TableTokenSecret: "table-secret-32-bytes-long-enough!",
		FrontendURL:      "http://localhost:5173",
	}
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestTableCreate_returnsValidToken(t *testing.T) {
	repo := newMockTableRepo()
	uc := usecase.NewTableUsecase(repo, testTableCfg())

	table, token, qrURL, err := uc.Create(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if table.TokenVersion != 1 {
		t.Errorf("want token_version=1, got %d", table.TokenVersion)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	if !strings.Contains(qrURL, "/table/") || !strings.Contains(qrURL, "?token=") {
		t.Errorf("qrURL shape unexpected: %s", qrURL)
	}

	// Token must verify correctly against the secret.
	claims, err := tabletoken.Verify(token, testTableCfg().TableTokenSecret)
	if err != nil {
		t.Fatalf("token should be valid: %v", err)
	}
	if claims.TableID != table.ID.String() {
		t.Errorf("tid mismatch: want %s got %s", table.ID, claims.TableID)
	}
	if claims.TokenVersion != 1 {
		t.Errorf("tvr should be 1, got %d", claims.TokenVersion)
	}
}

func TestTableRegenerateToken_bumpsVersionAndInvalidatesOldToken(t *testing.T) {
	repo := newMockTableRepo()
	uc := usecase.NewTableUsecase(repo, testTableCfg())

	table, oldToken, _, err := uc.Create(context.Background(), "2")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newToken, _, err := uc.RegenerateToken(context.Background(), table.ID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}

	// New token must carry version 2.
	newClaims, err := tabletoken.Verify(newToken, testTableCfg().TableTokenSecret)
	if err != nil {
		t.Fatalf("new token invalid: %v", err)
	}
	if newClaims.TokenVersion != 2 {
		t.Errorf("want tvr=2, got %d", newClaims.TokenVersion)
	}

	// Old token still verifies cryptographically but has tvr=1 which no longer matches table.token_version=2.
	oldClaims, err := tabletoken.Verify(oldToken, testTableCfg().TableTokenSecret)
	if err != nil {
		t.Fatalf("old token crypto invalid: %v", err)
	}
	currentTable, _ := repo.FindByID(context.Background(), table.ID)
	if oldClaims.TokenVersion == currentTable.TokenVersion {
		t.Error("old token tvr should not match current token_version after regeneration")
	}
}

func TestTableCreate_labelTaken(t *testing.T) {
	repo := newMockTableRepo()
	// Seed a table with label "3" manually to simulate unique constraint
	id := uuid.New()
	repo.tables[id] = &domain.Table{ID: id, Label: "3", TokenVersion: 1}

	// Mock doesn't enforce unique label — we test the domain error plumbing separately.
	// This test ensures Create flows without panicking when repo returns ErrTableLabelTaken.
	repo2 := &mockTableRepoWithLabelError{}
	uc := usecase.NewTableUsecase(repo2, testTableCfg())
	_, _, _, err := uc.Create(context.Background(), "3")
	if err == nil {
		t.Fatal("expected error for duplicate label")
	}
}

type mockTableRepoWithLabelError struct{ mockTableRepo }

func (m *mockTableRepoWithLabelError) Create(ctx context.Context, t *domain.Table) error {
	return domain.ErrTableLabelTaken
}

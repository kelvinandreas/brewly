package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/pkg/qrcode"
	"github.com/kelvinandreas/brewly/pkg/tabletoken"
)

// TableConfig holds secrets and URLs needed by TableUsecase.
type TableConfig struct {
	TableTokenSecret string
	FrontendURL      string
}

// TableUsecase handles business logic for cafe tables.
type TableUsecase struct {
	repo domain.TableRepository
	cfg  TableConfig
}

// NewTableUsecase constructs a TableUsecase.
func NewTableUsecase(repo domain.TableRepository, cfg TableConfig) *TableUsecase {
	return &TableUsecase{repo: repo, cfg: cfg}
}

// List returns all active tables.
func (u *TableUsecase) List(ctx context.Context) ([]domain.Table, error) {
	tables, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase.Table.List: %w", err)
	}
	return tables, nil
}

// Create inserts a new table, signs its first QR token, and returns the token + QR URL.
func (u *TableUsecase) Create(ctx context.Context, label string) (*domain.Table, string, string, error) {
	table := &domain.Table{
		ID:           uuid.New(),
		Label:        label,
		TokenVersion: 1,
	}
	if err := u.repo.Create(ctx, table); err != nil {
		return nil, "", "", fmt.Errorf("usecase.Table.Create: %w", err)
	}
	token, qrURL := u.signToken(table)
	return table, token, qrURL, nil
}

// Update patches the table label.
func (u *TableUsecase) Update(ctx context.Context, id uuid.UUID, label string) (*domain.Table, error) {
	table, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.Table.Update find: %w", err)
	}
	table.Label = label
	if err := u.repo.Update(ctx, table); err != nil {
		return nil, fmt.Errorf("usecase.Table.Update: %w", err)
	}
	return table, nil
}

// SoftDelete removes a table.
func (u *TableUsecase) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("usecase.Table.SoftDelete: %w", err)
	}
	return nil
}

// RegenerateToken bumps token_version (invalidating all existing tokens) and returns a new token + QR URL.
func (u *TableUsecase) RegenerateToken(ctx context.Context, id uuid.UUID) (string, string, error) {
	if err := u.repo.IncrementTokenVersion(ctx, id); err != nil {
		return "", "", fmt.Errorf("usecase.Table.RegenerateToken increment: %w", err)
	}
	table, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("usecase.Table.RegenerateToken find: %w", err)
	}
	token, qrURL := u.signToken(table)
	return token, qrURL, nil
}

// GetQRPNG returns a PNG QR code for the current table token.
func (u *TableUsecase) GetQRPNG(ctx context.Context, id uuid.UUID) ([]byte, error) {
	table, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.Table.GetQRPNG find: %w", err)
	}
	_, qrURL := u.signToken(table)
	png, err := qrcode.Generate(qrURL)
	if err != nil {
		return nil, fmt.Errorf("usecase.Table.GetQRPNG generate: %w", err)
	}
	return png, nil
}

// signToken signs a table token and builds the QR URL.
func (u *TableUsecase) signToken(table *domain.Table) (token, qrURL string) {
	token = tabletoken.Sign(table.ID, table.TokenVersion, u.cfg.TableTokenSecret)
	qrURL = fmt.Sprintf("%s/table/%s?token=%s", u.cfg.FrontendURL, table.ID, token)
	return
}

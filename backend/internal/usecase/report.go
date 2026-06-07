package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kelvinandreas/brewly/internal/domain"
)

// ReportUsecase handles reporting queries.
type ReportUsecase struct {
	repo domain.ReportRepository
}

// NewReportUsecase constructs a ReportUsecase.
func NewReportUsecase(repo domain.ReportRepository) *ReportUsecase {
	return &ReportUsecase{repo: repo}
}

// Revenue returns aggregated revenue bucketed by granularity within [from, to].
func (u *ReportUsecase) Revenue(ctx context.Context, granularity string, from, to time.Time) ([]domain.RevenueRow, error) {
	rows, err := u.repo.Revenue(ctx, granularity, from, to)
	if err != nil {
		return nil, fmt.Errorf("usecase.Report.Revenue: %w", err)
	}
	return rows, nil
}

// BestSellers returns top menu items by quantity in [from, to].
func (u *ReportUsecase) BestSellers(ctx context.Context, from, to time.Time, limit int) ([]domain.BestSellerRow, error) {
	rows, err := u.repo.BestSellers(ctx, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("usecase.Report.BestSellers: %w", err)
	}
	return rows, nil
}

// HourlyVolume returns order volume per hour for a single day.
func (u *ReportUsecase) HourlyVolume(ctx context.Context, date time.Time) ([]domain.HourlyVolumeRow, error) {
	rows, err := u.repo.HourlyVolume(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("usecase.Report.HourlyVolume: %w", err)
	}
	return rows, nil
}

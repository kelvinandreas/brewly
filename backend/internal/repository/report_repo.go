package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/your-handle/brewly/internal/domain"
	"gorm.io/gorm"
)

// ReportRepo implements domain.ReportRepository using GORM raw SQL.
type ReportRepo struct {
	db *gorm.DB
}

// NewReportRepo constructs a ReportRepo.
func NewReportRepo(db *gorm.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

// Revenue aggregates completed order totals by time bucket.
func (r *ReportRepo) Revenue(ctx context.Context, granularity string, from, to time.Time) ([]domain.RevenueRow, error) {
	// Validate granularity to prevent SQL injection.
	allowed := map[string]bool{"day": true, "week": true, "month": true}
	if !allowed[granularity] {
		return nil, fmt.Errorf("repository.ReportRepo.Revenue: invalid granularity %q", granularity)
	}

	type row struct {
		Period     time.Time
		TotalMinor int64
		OrderCount int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT date_trunc(?, created_at) AS period,
		       SUM(total_minor)          AS total_minor,
		       COUNT(*)                  AS order_count
		FROM orders
		WHERE status = 'completed'
		  AND created_at >= ?
		  AND created_at <= ?
		GROUP BY 1
		ORDER BY 1
	`, granularity, from, to).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.ReportRepo.Revenue: %w", err)
	}

	result := make([]domain.RevenueRow, len(rows))
	for i, row := range rows {
		result[i] = domain.RevenueRow{
			Period:     row.Period,
			TotalMinor: row.TotalMinor,
			OrderCount: row.OrderCount,
		}
	}
	return result, nil
}

// BestSellers returns the top-selling menu items by quantity in a date range.
func (r *ReportRepo) BestSellers(ctx context.Context, from, to time.Time, limit int) ([]domain.BestSellerRow, error) {
	type row struct {
		MenuItemID string
		Name       string
		TotalQty   int64
		TotalMinor int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT oi.menu_item_id::text AS menu_item_id,
		       oi.name_snapshot      AS name,
		       SUM(oi.quantity)      AS total_qty,
		       SUM(oi.quantity * oi.price_minor_snapshot) AS total_minor
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status = 'completed'
		  AND o.created_at >= ?
		  AND o.created_at <= ?
		GROUP BY oi.menu_item_id, oi.name_snapshot
		ORDER BY total_qty DESC
		LIMIT ?
	`, from, to, limit).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.ReportRepo.BestSellers: %w", err)
	}

	result := make([]domain.BestSellerRow, len(rows))
	for i, row := range rows {
		result[i] = domain.BestSellerRow{
			MenuItemID: row.MenuItemID,
			Name:       row.Name,
			TotalQty:   row.TotalQty,
			TotalMinor: row.TotalMinor,
		}
	}
	return result, nil
}

// HourlyVolume returns order count and revenue per hour for a single day.
func (r *ReportRepo) HourlyVolume(ctx context.Context, date time.Time) ([]domain.HourlyVolumeRow, error) {
	type row struct {
		Hour       int
		OrderCount int64
		TotalMinor int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXTRACT(HOUR FROM created_at)::int AS hour,
		       COUNT(*)                           AS order_count,
		       SUM(total_minor)                   AS total_minor
		FROM orders
		WHERE status = 'completed'
		  AND created_at::date = ?
		GROUP BY 1
		ORDER BY 1
	`, date.Format("2006-01-02")).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository.ReportRepo.HourlyVolume: %w", err)
	}

	result := make([]domain.HourlyVolumeRow, len(rows))
	for i, row := range rows {
		result[i] = domain.HourlyVolumeRow{
			Hour:       row.Hour,
			OrderCount: row.OrderCount,
			TotalMinor: row.TotalMinor,
		}
	}
	return result, nil
}

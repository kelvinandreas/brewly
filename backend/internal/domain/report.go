package domain

import (
	"context"
	"time"
)

// RevenueRow is one bucket in the revenue aggregation.
type RevenueRow struct {
	Period     time.Time `json:"period"`
	TotalMinor int64     `json:"totalMinor"`
	OrderCount int64     `json:"orderCount"`
}

// BestSellerRow holds aggregated sales data for a single menu item.
type BestSellerRow struct {
	MenuItemID string `json:"menuItemId"`
	Name       string `json:"name"`
	TotalQty   int64  `json:"totalQty"`
	TotalMinor int64  `json:"totalMinor"`
}

// HourlyVolumeRow holds order volume for a single hour of a day.
type HourlyVolumeRow struct {
	Hour       int   `json:"hour"`
	OrderCount int64 `json:"orderCount"`
	TotalMinor int64 `json:"totalMinor"`
}

// ReportRepository defines read-only aggregation queries for reports.
type ReportRepository interface {
	Revenue(ctx context.Context, granularity string, from, to time.Time) ([]RevenueRow, error)
	BestSellers(ctx context.Context, from, to time.Time, limit int) ([]BestSellerRow, error)
	HourlyVolume(ctx context.Context, date time.Time) ([]HourlyVolumeRow, error)
}

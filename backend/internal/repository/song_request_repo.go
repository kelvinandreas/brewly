package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"gorm.io/gorm"
)

type gormSongRequest struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	TableID        uuid.UUID `gorm:"type:uuid;not null"`
	TokenJTI       string    `gorm:"not null"`
	YoutubeVideoID string    `gorm:"not null"`
	Title          string    `gorm:"not null"`
	ChannelName    string    `gorm:"not null"`
	ThumbnailURL   string    `gorm:"not null"`
	Note           string
	Status         string    `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (gormSongRequest) TableName() string { return "song_requests" }

func songRequestToDomain(g *gormSongRequest) *domain.SongRequest {
	return &domain.SongRequest{
		ID:             g.ID,
		TableID:        g.TableID,
		TokenJTI:       g.TokenJTI,
		YoutubeVideoID: g.YoutubeVideoID,
		Title:          g.Title,
		ChannelName:    g.ChannelName,
		ThumbnailURL:   g.ThumbnailURL,
		Note:           g.Note,
		Status:         g.Status,
		CreatedAt:      g.CreatedAt,
		UpdatedAt:      g.UpdatedAt,
	}
}

// SongRequestRepo implements domain.SongRequestRepository using GORM.
type SongRequestRepo struct {
	db *gorm.DB
}

// NewSongRequestRepo constructs a SongRequestRepo.
func NewSongRequestRepo(db *gorm.DB) *SongRequestRepo {
	return &SongRequestRepo{db: db}
}

// Create inserts a new song request.
func (r *SongRequestRepo) Create(ctx context.Context, sr *domain.SongRequest) error {
	if sr.ID == uuid.Nil {
		sr.ID = uuid.New()
	}
	g := &gormSongRequest{
		ID:             sr.ID,
		TableID:        sr.TableID,
		TokenJTI:       sr.TokenJTI,
		YoutubeVideoID: sr.YoutubeVideoID,
		Title:          sr.Title,
		ChannelName:    sr.ChannelName,
		ThumbnailURL:   sr.ThumbnailURL,
		Note:           sr.Note,
		Status:         sr.Status,
	}
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("repository.SongRequestRepo.Create: %w", err)
	}
	sr.CreatedAt = g.CreatedAt
	sr.UpdatedAt = g.UpdatedAt
	return nil
}

// FindByID returns a single song request.
func (r *SongRequestRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SongRequest, error) {
	var g gormSongRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSongRequestNotFound
		}
		return nil, fmt.Errorf("repository.SongRequestRepo.FindByID: %w", err)
	}
	return songRequestToDomain(&g), nil
}

// List returns song requests, optionally filtered by status, newest first.
func (r *SongRequestRepo) List(ctx context.Context, status *string) ([]domain.SongRequest, error) {
	q := r.db.WithContext(ctx)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var rows []gormSongRequest
	if err := q.Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("repository.SongRequestRepo.List: %w", err)
	}
	result := make([]domain.SongRequest, len(rows))
	for i, g := range rows {
		gCopy := g
		result[i] = *songRequestToDomain(&gCopy)
	}
	return result, nil
}

// UpdateStatus sets the status on a single song request.
func (r *SongRequestRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).Model(&gormSongRequest{}).
		Where("id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("repository.SongRequestRepo.UpdateStatus: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrSongRequestNotFound
	}
	return nil
}

// CountActiveByJTI returns the number of queued requests for a given token JTI.
func (r *SongRequestRepo) CountActiveByJTI(ctx context.Context, jti string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&gormSongRequest{}).
		Where("token_jti = ? AND status = ?", jti, domain.SongQueued).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("repository.SongRequestRepo.CountActiveByJTI: %w", err)
	}
	return count, nil
}

// CountByStatus returns the number of requests with a given status.
func (r *SongRequestRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&gormSongRequest{}).
		Where("status = ?", status).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("repository.SongRequestRepo.CountByStatus: %w", err)
	}
	return count, nil
}

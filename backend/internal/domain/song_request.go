package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SongRequest represents a customer's song request.
type SongRequest struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TableID        uuid.UUID `gorm:"type:uuid;not null"                            json:"tableId"`
	TokenJTI       string    `gorm:"not null"                                      json:"tokenJti"`
	YoutubeVideoID string    `gorm:"not null"                                      json:"youtubeVideoId"`
	Title          string    `gorm:"not null"                                      json:"title"`
	ChannelName    string    `gorm:"not null"                                      json:"channelName"`
	ThumbnailURL   string    `gorm:"not null"                                      json:"thumbnailUrl"`
	Note           string    `json:"note"`
	Status         string    `gorm:"not null"                                      json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// SongRequestRepository defines persistence operations for song requests.
type SongRequestRepository interface {
	Create(ctx context.Context, sr *SongRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*SongRequest, error)
	List(ctx context.Context, status *string) ([]SongRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	CountActiveByJTI(ctx context.Context, jti string) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

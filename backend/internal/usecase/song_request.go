package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/pkg/sse"
)

// validSongTransitions maps current status → set of allowed next statuses.
var validSongTransitions = map[string]map[string]bool{
	domain.SongQueued:  {domain.SongPlaying: true, domain.SongSkipped: true},
	domain.SongPlaying: {domain.SongPlayed: true, domain.SongSkipped: true},
}

// SongRequestUsecase handles song request business logic.
type SongRequestUsecase struct {
	repo   domain.SongRequestRepository
	broker *sse.Broker
}

// NewSongRequestUsecase constructs a SongRequestUsecase.
func NewSongRequestUsecase(repo domain.SongRequestRepository, broker *sse.Broker) *SongRequestUsecase {
	return &SongRequestUsecase{repo: repo, broker: broker}
}

// Submit validates rate limit, creates a queued song request, and publishes SSE.
func (u *SongRequestUsecase) Submit(
	ctx context.Context,
	tableID uuid.UUID,
	tokenJTI string,
	videoID, title, channelName, thumbnailURL, note string,
) (*domain.SongRequest, error) {
	count, err := u.repo.CountActiveByJTI(ctx, tokenJTI)
	if err != nil {
		return nil, fmt.Errorf("usecase.SongRequest.Submit: %w", err)
	}
	if count >= domain.SongRequestRateLimit {
		return nil, domain.ErrSongRequestRateLimited
	}

	sr := &domain.SongRequest{
		TableID:        tableID,
		TokenJTI:       tokenJTI,
		YoutubeVideoID: videoID,
		Title:          title,
		ChannelName:    channelName,
		ThumbnailURL:   thumbnailURL,
		Note:           note,
		Status:         domain.SongQueued,
	}
	if err := u.repo.Create(ctx, sr); err != nil {
		return nil, fmt.Errorf("usecase.SongRequest.Submit: %w", err)
	}

	u.publishSong("song.requested", sr)
	return sr, nil
}

// List returns song requests, optionally filtered by status.
func (u *SongRequestUsecase) List(ctx context.Context, status *string) ([]domain.SongRequest, error) {
	items, err := u.repo.List(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("usecase.SongRequest.List: %w", err)
	}
	return items, nil
}

// UpdateStatus advances a song request status, enforcing the transition rules.
func (u *SongRequestUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string) (*domain.SongRequest, error) {
	sr, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.SongRequest.UpdateStatus: %w", err)
	}

	allowed, ok := validSongTransitions[sr.Status]
	if !ok || !allowed[newStatus] {
		return nil, domain.ErrInvalidSongStatusTransition
	}

	// Enforce single-playing invariant.
	if newStatus == domain.SongPlaying {
		playing, err := u.repo.CountByStatus(ctx, domain.SongPlaying)
		if err != nil {
			return nil, fmt.Errorf("usecase.SongRequest.UpdateStatus: %w", err)
		}
		if playing > 0 {
			return nil, domain.ErrSongAlreadyPlaying
		}
	}

	if err := u.repo.UpdateStatus(ctx, id, newStatus); err != nil {
		return nil, fmt.Errorf("usecase.SongRequest.UpdateStatus: %w", err)
	}
	sr.Status = newStatus
	u.publishSong("song.status_changed", sr)
	return sr, nil
}

func (u *SongRequestUsecase) publishSong(eventType string, sr *domain.SongRequest) {
	payload, _ := json.Marshal(sr)
	u.broker.Publish(sse.Event{Type: eventType, Payload: payload})
}

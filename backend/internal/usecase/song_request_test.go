package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/internal/usecase"
	"github.com/kelvinandreas/brewly/pkg/sse"
)

// ─── mock repository ──────────────────────────────────────────────────────────

type mockSongRepo struct {
	songs         map[uuid.UUID]*domain.SongRequest
	countByStatus map[string]int64
}

func newMockSongRepo() *mockSongRepo {
	return &mockSongRepo{
		songs:         make(map[uuid.UUID]*domain.SongRequest),
		countByStatus: make(map[string]int64),
	}
}

func (r *mockSongRepo) Create(_ context.Context, sr *domain.SongRequest) error {
	if sr.ID == uuid.Nil {
		sr.ID = uuid.New()
	}
	cp := *sr
	r.songs[sr.ID] = &cp
	r.countByStatus[sr.Status]++
	return nil
}

func (r *mockSongRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.SongRequest, error) {
	sr, ok := r.songs[id]
	if !ok {
		return nil, domain.ErrSongRequestNotFound
	}
	cp := *sr
	return &cp, nil
}

func (r *mockSongRepo) List(_ context.Context, status *string) ([]domain.SongRequest, error) {
	var list []domain.SongRequest
	for _, sr := range r.songs {
		if status == nil || sr.Status == *status {
			list = append(list, *sr)
		}
	}
	return list, nil
}

func (r *mockSongRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	sr, ok := r.songs[id]
	if !ok {
		return domain.ErrSongRequestNotFound
	}
	r.countByStatus[sr.Status]--
	sr.Status = status
	r.countByStatus[status]++
	return nil
}

func (r *mockSongRepo) CountActiveByJTI(_ context.Context, jti string) (int64, error) {
	var count int64
	for _, sr := range r.songs {
		if sr.TokenJTI == jti && sr.Status == domain.SongQueued {
			count++
		}
	}
	return count, nil
}

func (r *mockSongRepo) CountByStatus(_ context.Context, status string) (int64, error) {
	return r.countByStatus[status], nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newSongUC(repo *mockSongRepo) *usecase.SongRequestUsecase {
	return usecase.NewSongRequestUsecase(repo, sse.NewBroker())
}

func submitSong(t *testing.T, uc *usecase.SongRequestUsecase, jti string) *domain.SongRequest {
	t.Helper()
	sr, err := uc.Submit(context.Background(), uuid.New(), jti, "vid1", "Title", "Channel", "http://thumb", "")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	return sr
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestSongSubmit_success(t *testing.T) {
	repo := newMockSongRepo()
	uc := newSongUC(repo)

	sr, err := uc.Submit(context.Background(), uuid.New(), "jti-1", "vid1", "Jazz Vibes", "JazzChan", "http://t", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sr.Status != domain.SongQueued {
		t.Errorf("expected queued, got %s", sr.Status)
	}
}

func TestSongSubmit_rateLimited(t *testing.T) {
	repo := newMockSongRepo()
	uc := newSongUC(repo)
	jti := "jti-rate"

	for range domain.SongRequestRateLimit {
		submitSong(t, uc, jti)
	}

	_, err := uc.Submit(context.Background(), uuid.New(), jti, "v", "T", "C", "u", "")
	if !errors.Is(err, domain.ErrSongRequestRateLimited) {
		t.Errorf("expected ErrSongRequestRateLimited, got %v", err)
	}
}

func TestSongUpdateStatus_invalidTransition(t *testing.T) {
	repo := newMockSongRepo()
	uc := newSongUC(repo)

	sr := submitSong(t, uc, "jti-inv")

	// queued → played is not valid (must go through playing first)
	_, err := uc.UpdateStatus(context.Background(), sr.ID, domain.SongPlayed)
	if !errors.Is(err, domain.ErrInvalidSongStatusTransition) {
		t.Errorf("expected ErrInvalidSongStatusTransition, got %v", err)
	}
}

func TestSongUpdateStatus_doublePlay(t *testing.T) {
	repo := newMockSongRepo()
	uc := newSongUC(repo)

	sr1 := submitSong(t, uc, "jti-a")
	sr2 := submitSong(t, uc, "jti-b")

	if _, err := uc.UpdateStatus(context.Background(), sr1.ID, domain.SongPlaying); err != nil {
		t.Fatalf("first play: %v", err)
	}

	_, err := uc.UpdateStatus(context.Background(), sr2.ID, domain.SongPlaying)
	if !errors.Is(err, domain.ErrSongAlreadyPlaying) {
		t.Errorf("expected ErrSongAlreadyPlaying, got %v", err)
	}
}

func TestSongUpdateStatus_validChain(t *testing.T) {
	repo := newMockSongRepo()
	uc := newSongUC(repo)

	sr := submitSong(t, uc, "jti-chain")

	sr2, err := uc.UpdateStatus(context.Background(), sr.ID, domain.SongPlaying)
	if err != nil {
		t.Fatalf("play: %v", err)
	}
	if sr2.Status != domain.SongPlaying {
		t.Errorf("expected playing, got %s", sr2.Status)
	}

	sr3, err := uc.UpdateStatus(context.Background(), sr.ID, domain.SongPlayed)
	if err != nil {
		t.Fatalf("played: %v", err)
	}
	if sr3.Status != domain.SongPlayed {
		t.Errorf("expected played, got %s", sr3.Status)
	}
}

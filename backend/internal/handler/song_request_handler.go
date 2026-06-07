package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/response"
)

// SongRequestHandler handles /api/song-requests endpoints.
type SongRequestHandler struct {
	uc *usecase.SongRequestUsecase
}

// NewSongRequestHandler constructs a SongRequestHandler.
func NewSongRequestHandler(uc *usecase.SongRequestUsecase) *SongRequestHandler {
	return &SongRequestHandler{uc: uc}
}

// List GET /api/song-requests?status=
func (h *SongRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	items, err := h.uc.List(r.Context(), status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not list song requests")
		return
	}
	response.OK(w, map[string]any{"songRequests": items})
}

// UpdateStatus PATCH /api/song-requests/:id/status
func (h *SongRequestHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	sr, err := h.uc.UpdateStatus(r.Context(), id, body.Status)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSongRequestNotFound):
			response.Error(w, http.StatusNotFound, "song_not_found", "song request not found")
		case errors.Is(err, domain.ErrInvalidSongStatusTransition):
			response.Error(w, http.StatusUnprocessableEntity, "invalid_transition", "invalid status transition")
		case errors.Is(err, domain.ErrSongAlreadyPlaying):
			response.Error(w, http.StatusConflict, "song_already_playing", "another song is already playing")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not update status")
		}
		return
	}
	response.OK(w, map[string]any{"songRequest": sr})
}

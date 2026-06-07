package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/internal/usecase"
	"github.com/kelvinandreas/brewly/pkg/response"
)

// UserHandler handles owner-only /api/users endpoints.
type UserHandler struct {
	uc *usecase.UserUsecase
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

// List GET /api/users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.uc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not list users")
		return
	}
	response.OK(w, map[string]any{"users": users})
}

// Create POST /api/users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	user, err := h.uc.Create(r.Context(), body.Email, body.Password, body.Name, body.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailTaken):
			response.Error(w, http.StatusConflict, "email_taken", "email already in use")
		case errors.Is(err, domain.ErrForbidden):
			response.Error(w, http.StatusForbidden, "forbidden", "cannot create owner via this endpoint")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not create user")
		}
		return
	}
	response.Created(w, map[string]any{"user": user})
}

// Update PATCH /api/users/:id
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid user id")
		return
	}

	var body struct {
		Name     *string `json:"name"`
		Role     *string `json:"role"`
		Password *string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	user, err := h.uc.UpdateFields(r.Context(), id, body.Name, body.Role, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			response.Error(w, http.StatusNotFound, "user_not_found", "user not found")
		case errors.Is(err, domain.ErrForbidden):
			response.Error(w, http.StatusForbidden, "forbidden", "cannot assign owner role")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not update user")
		}
		return
	}
	response.OK(w, map[string]any{"user": user})
}

// Delete DELETE /api/users/:id
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid user id")
		return
	}

	if err := h.uc.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user_not_found", "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not delete user")
		return
	}
	response.NoContent(w)
}

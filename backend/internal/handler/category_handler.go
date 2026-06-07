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

// CategoryHandler handles /api/categories endpoints.
type CategoryHandler struct {
	uc *usecase.CategoryUsecase
}

// NewCategoryHandler constructs a CategoryHandler.
func NewCategoryHandler(uc *usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

// List GET /api/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.uc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not list categories")
		return
	}
	response.OK(w, map[string]any{"categories": cats})
}

// Create POST /api/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string `json:"name"`
		DisplayOrder int    `json:"displayOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if body.Name == "" {
		response.ValidationError(w, []response.FieldError{{Field: "name", Message: "required"}})
		return
	}
	cat, err := h.uc.Create(r.Context(), body.Name, body.DisplayOrder)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not create category")
		return
	}
	response.Created(w, map[string]any{"category": cat})
}

// Update PATCH /api/categories/:id
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var body struct {
		Name         *string `json:"name"`
		DisplayOrder *int    `json:"displayOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	cat, err := h.uc.Update(r.Context(), id, body.Name, body.DisplayOrder)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			response.Error(w, http.StatusNotFound, "category_not_found", "category not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not update category")
		return
	}
	response.OK(w, map[string]any{"category": cat})
}

// Delete DELETE /api/categories/:id
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := h.uc.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			response.Error(w, http.StatusNotFound, "category_not_found", "category not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not delete category")
		return
	}
	response.NoContent(w)
}

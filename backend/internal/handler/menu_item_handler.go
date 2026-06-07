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

// MenuItemHandler handles /api/menu-items endpoints.
type MenuItemHandler struct {
	uc *usecase.MenuItemUsecase
}

// NewMenuItemHandler constructs a MenuItemHandler.
func NewMenuItemHandler(uc *usecase.MenuItemUsecase) *MenuItemHandler {
	return &MenuItemHandler{uc: uc}
}

// List GET /api/menu-items?categoryId=&availableOnly=
func (h *MenuItemHandler) List(w http.ResponseWriter, r *http.Request) {
	var catID *uuid.UUID
	if raw := r.URL.Query().Get("categoryId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "bad_request", "invalid categoryId")
			return
		}
		catID = &id
	}
	availableOnly := r.URL.Query().Get("availableOnly") == "true"

	items, err := h.uc.List(r.Context(), catID, availableOnly)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not list menu items")
		return
	}
	response.OK(w, map[string]any{"items": items})
}

// Create POST /api/menu-items
func (h *MenuItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CategoryID  string  `json:"categoryId"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		PriceMinor  int64   `json:"priceMinor"`
		ImageURL    *string `json:"imageUrl"`
		IsAvailable *bool   `json:"isAvailable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	catID, err := uuid.Parse(body.CategoryID)
	if err != nil {
		response.ValidationError(w, []response.FieldError{{Field: "categoryId", Message: "valid UUID required"}})
		return
	}
	isAvailable := true
	if body.IsAvailable != nil {
		isAvailable = *body.IsAvailable
	}
	item, err := h.uc.Create(r.Context(), domain.MenuItem{
		CategoryID:  catID,
		Name:        body.Name,
		Description: body.Description,
		PriceMinor:  body.PriceMinor,
		ImageURL:    body.ImageURL,
		IsAvailable: isAvailable,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			response.Error(w, http.StatusUnprocessableEntity, "category_not_found", "category does not exist")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not create menu item")
		return
	}
	response.Created(w, map[string]any{"item": item})
}

// Update PATCH /api/menu-items/:id
func (h *MenuItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var body struct {
		CategoryID  *string `json:"categoryId"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		PriceMinor  int64   `json:"priceMinor"`
		ImageURL    *string `json:"imageUrl"`
		IsAvailable bool    `json:"isAvailable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	patch := domain.MenuItem{
		Name:        body.Name,
		Description: body.Description,
		PriceMinor:  body.PriceMinor,
		ImageURL:    body.ImageURL,
		IsAvailable: body.IsAvailable,
	}
	if body.CategoryID != nil {
		cid, err := uuid.Parse(*body.CategoryID)
		if err != nil {
			response.ValidationError(w, []response.FieldError{{Field: "categoryId", Message: "valid UUID required"}})
			return
		}
		patch.CategoryID = cid
	}
	item, err := h.uc.Update(r.Context(), id, patch)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMenuItemNotFound):
			response.Error(w, http.StatusNotFound, "item_not_found", "menu item not found")
		case errors.Is(err, domain.ErrCategoryNotFound):
			response.Error(w, http.StatusUnprocessableEntity, "category_not_found", "category does not exist")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not update menu item")
		}
		return
	}
	response.OK(w, map[string]any{"item": item})
}

// Delete DELETE /api/menu-items/:id
func (h *MenuItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := h.uc.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrMenuItemNotFound) {
			response.Error(w, http.StatusNotFound, "item_not_found", "menu item not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not delete menu item")
		return
	}
	response.NoContent(w)
}

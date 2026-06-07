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

// TableHandler handles /api/tables endpoints.
type TableHandler struct {
	uc *usecase.TableUsecase
}

// NewTableHandler constructs a TableHandler.
func NewTableHandler(uc *usecase.TableUsecase) *TableHandler {
	return &TableHandler{uc: uc}
}

// List GET /api/tables
func (h *TableHandler) List(w http.ResponseWriter, r *http.Request) {
	tables, err := h.uc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not list tables")
		return
	}
	response.OK(w, map[string]any{"tables": tables})
}

// Create POST /api/tables
func (h *TableHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if body.Label == "" {
		response.ValidationError(w, []response.FieldError{{Field: "label", Message: "required"}})
		return
	}
	table, token, qrURL, err := h.uc.Create(r.Context(), body.Label)
	if err != nil {
		if errors.Is(err, domain.ErrTableLabelTaken) {
			response.Error(w, http.StatusConflict, "label_taken", "table label already in use")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not create table")
		return
	}
	response.Created(w, map[string]any{
		"table":   table,
		"qrToken": token,
		"qrUrl":   qrURL,
	})
}

// Update PATCH /api/tables/:id
func (h *TableHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var body struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	table, err := h.uc.Update(r.Context(), id, body.Label)
	if err != nil {
		if errors.Is(err, domain.ErrTableNotFound) {
			response.Error(w, http.StatusNotFound, "table_not_found", "table not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not update table")
		return
	}
	response.OK(w, map[string]any{"table": table})
}

// Delete DELETE /api/tables/:id
func (h *TableHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := h.uc.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrTableNotFound) {
			response.Error(w, http.StatusNotFound, "table_not_found", "table not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not delete table")
		return
	}
	response.NoContent(w)
}

// RegenerateToken POST /api/tables/:id/regenerate-token
func (h *TableHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	token, qrURL, err := h.uc.RegenerateToken(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTableNotFound) {
			response.Error(w, http.StatusNotFound, "table_not_found", "table not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not regenerate token")
		return
	}
	response.OK(w, map[string]any{"qrToken": token, "qrUrl": qrURL})
}

// GetQR GET /api/tables/:id/qr.png
func (h *TableHandler) GetQR(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	png, err := h.uc.GetQRPNG(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTableNotFound) {
			response.Error(w, http.StatusNotFound, "table_not_found", "table not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not generate QR code")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/middleware"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/response"
)

// OrderHandler handles /api/orders endpoints.
type OrderHandler struct {
	uc *usecase.OrderUsecase
}

// NewOrderHandler constructs an OrderHandler.
func NewOrderHandler(uc *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

// List GET /api/orders?status=&from=&to=
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	var filter domain.OrderFilter
	if s := r.URL.Query().Get("status"); s != "" {
		filter.Status = &s
	}
	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "bad_request", "invalid from date (RFC3339 required)")
			return
		}
		filter.From = &t
	}
	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "bad_request", "invalid to date (RFC3339 required)")
			return
		}
		filter.To = &t
	}
	orders, err := h.uc.List(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not list orders")
		return
	}
	response.OK(w, map[string]any{"orders": orders})
}

// GetByID GET /api/orders/:id
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	order, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.Error(w, http.StatusNotFound, "order_not_found", "order not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not get order")
		return
	}
	response.OK(w, map[string]any{"order": order})
}

// Create POST /api/orders (cashier/owner)
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TableID string `json:"tableId"`
		Items   []struct {
			MenuItemID string `json:"menuItemId"`
			Quantity   int    `json:"quantity"`
		} `json:"items"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	tableID, err := uuid.Parse(body.TableID)
	if err != nil {
		response.ValidationError(w, []response.FieldError{{Field: "tableId", Message: "valid UUID required"}})
		return
	}
	userIDStr := middleware.UserIDFromCtx(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid user context")
		return
	}

	inputs := make([]domain.OrderItemInput, len(body.Items))
	for i, it := range body.Items {
		mid, err := uuid.Parse(it.MenuItemID)
		if err != nil {
			response.ValidationError(w, []response.FieldError{{Field: "items[" + string(rune('0'+i)) + "].menuItemId", Message: "valid UUID required"}})
			return
		}
		inputs[i] = domain.OrderItemInput{MenuItemID: mid, Quantity: it.Quantity}
	}

	order, err := h.uc.CreateForCashier(r.Context(), tableID, userID, inputs, body.Note)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMenuItemNotFound):
			response.Error(w, http.StatusUnprocessableEntity, "item_not_found", "menu item not found")
		case errors.Is(err, domain.ErrMenuItemUnavailable):
			response.Error(w, http.StatusUnprocessableEntity, "item_unavailable", "menu item is not available")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not create order")
		}
		return
	}
	response.Created(w, map[string]any{"order": order})
}

// AdvanceStatus PATCH /api/orders/:id/status
func (h *OrderHandler) AdvanceStatus(w http.ResponseWriter, r *http.Request) {
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
	order, err := h.uc.AdvanceStatus(r.Context(), id, body.Status)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			response.Error(w, http.StatusNotFound, "order_not_found", "order not found")
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			response.Error(w, http.StatusUnprocessableEntity, "invalid_transition", "invalid status transition")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not update status")
		}
		return
	}
	response.OK(w, map[string]any{"order": order})
}

// Cancel POST /api/orders/:id/cancel
func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	order, err := h.uc.Cancel(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			response.Error(w, http.StatusNotFound, "order_not_found", "order not found")
		case errors.Is(err, domain.ErrOrderCancelled):
			response.Error(w, http.StatusUnprocessableEntity, "order_cancelled", "order cannot be cancelled")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not cancel order")
		}
		return
	}
	response.OK(w, map[string]any{"order": order})
}

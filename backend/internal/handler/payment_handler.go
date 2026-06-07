package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/internal/middleware"
	"github.com/kelvinandreas/brewly/internal/usecase"
	"github.com/kelvinandreas/brewly/pkg/response"
)

// PaymentHandler handles /api/orders/:id/payments endpoints.
type PaymentHandler struct {
	uc *usecase.PaymentUsecase
}

// NewPaymentHandler constructs a PaymentHandler.
func NewPaymentHandler(uc *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

// Create POST /api/orders/:id/payments
func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid order id")
		return
	}
	var body struct {
		Method        string `json:"method"`
		AmountMinor   int64  `json:"amountMinor"`
		ReceivedMinor int64  `json:"receivedMinor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	userIDStr := middleware.UserIDFromCtx(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid user context")
		return
	}

	payment, err := h.uc.Record(r.Context(), orderID, userID, body.Method, body.AmountMinor, body.ReceivedMinor)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.Error(w, http.StatusNotFound, "order_not_found", "order not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not record payment")
		return
	}
	response.Created(w, map[string]any{"payment": payment})
}

// List GET /api/orders/:id/payments
func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid order id")
		return
	}
	payments, err := h.uc.ListByOrder(r.Context(), orderID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not list payments")
		return
	}
	response.OK(w, map[string]any{"payments": payments})
}

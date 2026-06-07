package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/middleware"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/response"
)

// CustomerHandler handles unauthenticated customer endpoints (protected by table token).
type CustomerHandler struct {
	catUC   *usecase.CategoryUsecase
	itemUC  *usecase.MenuItemUsecase
	orderUC *usecase.OrderUsecase
}

// NewCustomerHandler constructs a CustomerHandler.
func NewCustomerHandler(catUC *usecase.CategoryUsecase, itemUC *usecase.MenuItemUsecase, orderUC *usecase.OrderUsecase) *CustomerHandler {
	return &CustomerHandler{catUC: catUC, itemUC: itemUC, orderUC: orderUC}
}

// customerMenuItem is the public-facing shape for a customer (no internal fields).
type customerMenuItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	PriceMinor  int64   `json:"priceMinor"`
	ImageURL    *string `json:"imageUrl"`
}

type customerMenuCategory struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	DisplayOrder int                `json:"displayOrder"`
	Items        []customerMenuItem `json:"items"`
}

// GetMenu GET /api/customer/menu
// Returns categories ordered by display_order, each with only available items.
func (h *CustomerHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	cats, err := h.catUC.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not load menu")
		return
	}

	availableOnly := true
	result := make([]customerMenuCategory, 0, len(cats))
	for _, cat := range cats {
		catID := cat.ID
		items, err := h.itemUC.List(r.Context(), &catID, availableOnly)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not load menu items")
			return
		}
		menuItems := make([]customerMenuItem, len(items))
		for i, item := range items {
			menuItems[i] = toCustomerItem(item)
		}
		result = append(result, customerMenuCategory{
			ID:           cat.ID.String(),
			Name:         cat.Name,
			DisplayOrder: cat.DisplayOrder,
			Items:        menuItems,
		})
	}
	response.OK(w, map[string]any{"categories": result})
}

func toCustomerItem(m domain.MenuItem) customerMenuItem {
	return customerMenuItem{
		ID:          m.ID.String(),
		Name:        m.Name,
		Description: m.Description,
		PriceMinor:  m.PriceMinor,
		ImageURL:    m.ImageURL,
	}
}

// PlaceOrder POST /api/customer/orders
func (h *CustomerHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	tableIDStr := middleware.TableIDFromCtx(r.Context())
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid table context")
		return
	}

	var body struct {
		Items []struct {
			MenuItemID string `json:"menuItemId"`
			Quantity   int    `json:"quantity"`
		} `json:"items"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	inputs := make([]domain.OrderItemInput, len(body.Items))
	for i, it := range body.Items {
		mid, err := uuid.Parse(it.MenuItemID)
		if err != nil {
			response.ValidationError(w, []response.FieldError{{Field: "items[].menuItemId", Message: "valid UUID required"}})
			return
		}
		inputs[i] = domain.OrderItemInput{MenuItemID: mid, Quantity: it.Quantity}
	}

	order, err := h.orderUC.CreateForTable(r.Context(), tableID, inputs, body.Note)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMenuItemNotFound):
			response.Error(w, http.StatusUnprocessableEntity, "item_not_found", "menu item not found")
		case errors.Is(err, domain.ErrMenuItemUnavailable):
			response.Error(w, http.StatusUnprocessableEntity, "item_unavailable", "menu item is not available")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not place order")
		}
		return
	}
	response.Created(w, map[string]any{"order": order})
}

// MyOrders GET /api/customer/orders/mine
func (h *CustomerHandler) MyOrders(w http.ResponseWriter, r *http.Request) {
	tableIDStr := middleware.TableIDFromCtx(r.Context())
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid table context")
		return
	}
	orders, err := h.orderUC.ListByTable(r.Context(), tableID, 5)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not load orders")
		return
	}
	response.OK(w, map[string]any{"orders": orders})
}

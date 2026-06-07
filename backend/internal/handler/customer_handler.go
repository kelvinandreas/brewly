package handler

import (
	"net/http"

	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/response"
)

// CustomerHandler handles unauthenticated customer endpoints (protected by table token).
type CustomerHandler struct {
	catUC  *usecase.CategoryUsecase
	itemUC *usecase.MenuItemUsecase
}

// NewCustomerHandler constructs a CustomerHandler.
func NewCustomerHandler(catUC *usecase.CategoryUsecase, itemUC *usecase.MenuItemUsecase) *CustomerHandler {
	return &CustomerHandler{catUC: catUC, itemUC: itemUC}
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

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/middleware"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/response"
	"github.com/your-handle/brewly/pkg/youtube"
)

// youtubeClient is a local interface so we don't import pkg/youtube in handler tests.
type youtubeClient interface {
	Search(ctx context.Context, q string, maxResults int) ([]youtube.VideoResult, error)
}

// CustomerHandler handles unauthenticated customer endpoints (protected by table token).
type CustomerHandler struct {
	catUC    *usecase.CategoryUsecase
	itemUC   *usecase.MenuItemUsecase
	orderUC  *usecase.OrderUsecase
	songUC   *usecase.SongRequestUsecase
	ytClient youtubeClient
}

// NewCustomerHandler constructs a CustomerHandler.
func NewCustomerHandler(
	catUC *usecase.CategoryUsecase,
	itemUC *usecase.MenuItemUsecase,
	orderUC *usecase.OrderUsecase,
	songUC *usecase.SongRequestUsecase,
	yt youtubeClient,
) *CustomerHandler {
	return &CustomerHandler{catUC: catUC, itemUC: itemUC, orderUC: orderUC, songUC: songUC, ytClient: yt}
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

// YouTubeSearch GET /api/customer/songs/search?q=<query>&maxResults=<n>
func (h *CustomerHandler) YouTubeSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "q is required")
		return
	}
	maxResults := 10
	if mr := r.URL.Query().Get("maxResults"); mr != "" {
		n, err := strconv.Atoi(mr)
		if err != nil || n <= 0 {
			response.Error(w, http.StatusBadRequest, "bad_request", "invalid maxResults")
			return
		}
		maxResults = n
	}

	results, err := h.ytClient.Search(r.Context(), q, maxResults)
	if err != nil {
		if errors.Is(err, youtube.ErrKeyNotConfigured) {
			response.Error(w, http.StatusNotImplemented, "not_implemented", "YouTube search is not configured")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not search YouTube")
		return
	}
	response.OK(w, map[string]any{"results": results})
}

// SubmitSongRequest POST /api/customer/songs
func (h *CustomerHandler) SubmitSongRequest(w http.ResponseWriter, r *http.Request) {
	tableIDStr := middleware.TableIDFromCtx(r.Context())
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid table context")
		return
	}
	tokenJTI := middleware.TokenJTIFromCtx(r.Context())
	if tokenJTI == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid table context")
		return
	}

	var body struct {
		VideoID      string `json:"videoId"`
		Title        string `json:"title"`
		ChannelName  string `json:"channelName"`
		ThumbnailURL string `json:"thumbnailUrl"`
		Note         string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if body.VideoID == "" || body.Title == "" {
		response.ValidationError(w, []response.FieldError{
			{Field: "videoId", Message: "required"},
			{Field: "title", Message: "required"},
		})
		return
	}

	sr, err := h.songUC.Submit(r.Context(), tableID, tokenJTI, body.VideoID, body.Title, body.ChannelName, body.ThumbnailURL, body.Note)
	if err != nil {
		if errors.Is(err, domain.ErrSongRequestRateLimited) {
			response.Error(w, http.StatusTooManyRequests, "rate_limited", "you have too many queued song requests")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not submit song request")
		return
	}
	response.Created(w, map[string]any{"songRequest": sr})
}

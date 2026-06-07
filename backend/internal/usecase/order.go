package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/pkg/sse"
)

// validTransitions maps each non-terminal status to the only valid next status.
var validTransitions = map[string]string{
	domain.StatusPending:   domain.StatusConfirmed,
	domain.StatusConfirmed: domain.StatusPreparing,
	domain.StatusPreparing: domain.StatusReady,
	domain.StatusReady:     domain.StatusCompleted,
}

// terminalStatuses are statuses from which no further transitions are allowed.
var terminalStatuses = map[string]bool{
	domain.StatusCompleted: true,
	domain.StatusCancelled: true,
}

// OrderUsecase handles order business logic.
type OrderUsecase struct {
	orderRepo    domain.OrderRepository
	menuItemRepo domain.MenuItemRepository
	broker       *sse.Broker
}

// NewOrderUsecase constructs an OrderUsecase.
func NewOrderUsecase(orderRepo domain.OrderRepository, menuItemRepo domain.MenuItemRepository, broker *sse.Broker) *OrderUsecase {
	return &OrderUsecase{orderRepo: orderRepo, menuItemRepo: menuItemRepo, broker: broker}
}

// CreateForTable places an order from a customer QR scan.
func (u *OrderUsecase) CreateForTable(ctx context.Context, tableID uuid.UUID, inputs []domain.OrderItemInput, note string) (*domain.Order, error) {
	return u.createOrder(ctx, tableID, nil, domain.SourceCustomerQR, inputs, note)
}

// CreateForCashier places an order from the cashier POS.
func (u *OrderUsecase) CreateForCashier(ctx context.Context, tableID, createdByUserID uuid.UUID, inputs []domain.OrderItemInput, note string) (*domain.Order, error) {
	return u.createOrder(ctx, tableID, &createdByUserID, domain.SourceCashierPOS, inputs, note)
}

func (u *OrderUsecase) createOrder(
	ctx context.Context,
	tableID uuid.UUID,
	createdByUserID *uuid.UUID,
	source string,
	inputs []domain.OrderItemInput,
	note string,
) (*domain.Order, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("usecase.createOrder: at least one item required")
	}

	var items []domain.OrderItem
	var totalMinor int64

	for _, inp := range inputs {
		if inp.Quantity <= 0 {
			return nil, fmt.Errorf("usecase.createOrder: quantity must be positive")
		}
		mi, err := u.menuItemRepo.FindByID(ctx, inp.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("usecase.createOrder: %w", err)
		}
		if !mi.IsAvailable {
			return nil, domain.ErrMenuItemUnavailable
		}
		items = append(items, domain.OrderItem{
			MenuItemID:         mi.ID,
			NameSnapshot:       mi.Name,
			PriceMinorSnapshot: mi.PriceMinor,
			Quantity:           inp.Quantity,
		})
		totalMinor += mi.PriceMinor * int64(inp.Quantity)
	}

	order := &domain.Order{
		TableID:         tableID,
		Status:          domain.StatusPending,
		Source:          source,
		TotalMinor:      totalMinor,
		Note:            note,
		CreatedByUserID: createdByUserID,
		Items:           items,
	}

	if err := u.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("usecase.createOrder: %w", err)
	}

	u.publishOrder("order.created", order)
	return order, nil
}

// List returns orders matching the filter.
func (u *OrderUsecase) List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, error) {
	orders, err := u.orderRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("usecase.OrderUsecase.List: %w", err)
	}
	return orders, nil
}

// GetByID returns a single order by ID.
func (u *OrderUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.OrderUsecase.GetByID: %w", err)
	}
	return order, nil
}

// AdvanceStatus moves an order forward one step in the state machine.
func (u *OrderUsecase) AdvanceStatus(ctx context.Context, id uuid.UUID, newStatus string) (*domain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.AdvanceStatus: %w", err)
	}

	expected, ok := validTransitions[order.Status]
	if !ok || expected != newStatus {
		return nil, domain.ErrInvalidStatusTransition
	}

	if err := u.orderRepo.UpdateStatus(ctx, id, newStatus); err != nil {
		return nil, fmt.Errorf("usecase.AdvanceStatus: %w", err)
	}
	order.Status = newStatus
	u.publishOrder("order.status_changed", order)
	return order, nil
}

// Cancel moves an order to cancelled from any non-terminal state.
func (u *OrderUsecase) Cancel(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.Cancel: %w", err)
	}
	if terminalStatuses[order.Status] {
		return nil, domain.ErrOrderCancelled
	}
	if err := u.orderRepo.UpdateStatus(ctx, id, domain.StatusCancelled); err != nil {
		return nil, fmt.Errorf("usecase.Cancel: %w", err)
	}
	order.Status = domain.StatusCancelled
	u.publishOrder("order.cancelled", order)
	return order, nil
}

// ListByTable returns the most recent orders for a given table.
func (u *OrderUsecase) ListByTable(ctx context.Context, tableID uuid.UUID, limit int) ([]domain.Order, error) {
	orders, err := u.orderRepo.ListByTable(ctx, tableID, limit)
	if err != nil {
		return nil, fmt.Errorf("usecase.OrderUsecase.ListByTable: %w", err)
	}
	return orders, nil
}

func (u *OrderUsecase) publishOrder(eventType string, order *domain.Order) {
	payload, _ := json.Marshal(order)
	u.broker.Publish(sse.Event{Type: eventType, Payload: payload})
}

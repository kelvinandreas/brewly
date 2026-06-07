package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/internal/usecase"
	"github.com/kelvinandreas/brewly/pkg/sse"
)

// ─── mock repositories ────────────────────────────────────────────────────────

type mockOrderRepo struct {
	orders      map[uuid.UUID]*domain.Order
	createErr   error
	findByIDErr error
	updateErr   error
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[uuid.UUID]*domain.Order)}
}

func (r *mockOrderRepo) Create(_ context.Context, o *domain.Order) error {
	if r.createErr != nil {
		return r.createErr
	}
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	cp := *o
	r.orders[o.ID] = &cp
	return nil
}

func (r *mockOrderRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Order, error) {
	if r.findByIDErr != nil {
		return nil, r.findByIDErr
	}
	o, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *mockOrderRepo) List(_ context.Context, _ domain.OrderFilter) ([]domain.Order, error) {
	var list []domain.Order
	for _, o := range r.orders {
		list = append(list, *o)
	}
	return list, nil
}

func (r *mockOrderRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	o, ok := r.orders[id]
	if !ok {
		return domain.ErrOrderNotFound
	}
	o.Status = status
	return nil
}

func (r *mockOrderRepo) ListByTable(_ context.Context, tableID uuid.UUID, limit int) ([]domain.Order, error) {
	var list []domain.Order
	for _, o := range r.orders {
		if o.TableID == tableID {
			list = append(list, *o)
		}
		if len(list) >= limit {
			break
		}
	}
	return list, nil
}

type mockMenuItemRepoForOrder struct {
	items   map[uuid.UUID]*domain.MenuItem
	findErr error
}

func newMockMenuItemRepoForOrder() *mockMenuItemRepoForOrder {
	return &mockMenuItemRepoForOrder{items: make(map[uuid.UUID]*domain.MenuItem)}
}

func (r *mockMenuItemRepoForOrder) FindByID(_ context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	m, ok := r.items[id]
	if !ok {
		return nil, domain.ErrMenuItemNotFound
	}
	cp := *m
	return &cp, nil
}

// Satisfy the full interface with no-op stubs.
func (r *mockMenuItemRepoForOrder) List(_ context.Context, _ domain.MenuItemFilter) ([]domain.MenuItem, error) {
	return nil, nil
}
func (r *mockMenuItemRepoForOrder) Create(_ context.Context, _ *domain.MenuItem) error { return nil }
func (r *mockMenuItemRepoForOrder) Update(_ context.Context, _ *domain.MenuItem) error { return nil }
func (r *mockMenuItemRepoForOrder) SoftDelete(_ context.Context, _ uuid.UUID) error    { return nil }

// ─── helpers ──────────────────────────────────────────────────────────────────

func newOrderUC(orderRepo *mockOrderRepo, menuRepo *mockMenuItemRepoForOrder) *usecase.OrderUsecase {
	broker := sse.NewBroker()
	return usecase.NewOrderUsecase(orderRepo, menuRepo, broker)
}

func seedMenuItem(repo *mockMenuItemRepoForOrder, available bool) *domain.MenuItem {
	m := &domain.MenuItem{
		ID:          uuid.New(),
		Name:        "Espresso",
		PriceMinor:  25000,
		IsAvailable: available,
	}
	repo.items[m.ID] = m
	return m
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestCreateForTable_success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	menuRepo := newMockMenuItemRepoForOrder()
	item := seedMenuItem(menuRepo, true)
	uc := newOrderUC(orderRepo, menuRepo)

	tableID := uuid.New()
	order, err := uc.CreateForTable(context.Background(), tableID, []domain.OrderItemInput{
		{MenuItemID: item.ID, Quantity: 2},
	}, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.TotalMinor != 50000 {
		t.Errorf("expected total 50000, got %d", order.TotalMinor)
	}
	if order.Status != domain.StatusPending {
		t.Errorf("expected status pending, got %s", order.Status)
	}
	if order.Source != domain.SourceCustomerQR {
		t.Errorf("expected source customer_qr, got %s", order.Source)
	}
}

func TestCreateForTable_unavailableItem(t *testing.T) {
	orderRepo := newMockOrderRepo()
	menuRepo := newMockMenuItemRepoForOrder()
	item := seedMenuItem(menuRepo, false)
	uc := newOrderUC(orderRepo, menuRepo)

	_, err := uc.CreateForTable(context.Background(), uuid.New(), []domain.OrderItemInput{
		{MenuItemID: item.ID, Quantity: 1},
	}, "")

	if !errors.Is(err, domain.ErrMenuItemUnavailable) {
		t.Errorf("expected ErrMenuItemUnavailable, got %v", err)
	}
}

func TestCreateForTable_itemNotFound(t *testing.T) {
	uc := newOrderUC(newMockOrderRepo(), newMockMenuItemRepoForOrder())

	_, err := uc.CreateForTable(context.Background(), uuid.New(), []domain.OrderItemInput{
		{MenuItemID: uuid.New(), Quantity: 1},
	}, "")

	if !errors.Is(err, domain.ErrMenuItemNotFound) {
		t.Errorf("expected ErrMenuItemNotFound, got %v", err)
	}
}

func TestAdvanceStatus_validChain(t *testing.T) {
	chain := []string{
		domain.StatusConfirmed,
		domain.StatusPreparing,
		domain.StatusReady,
		domain.StatusCompleted,
	}

	orderRepo := newMockOrderRepo()
	menuRepo := newMockMenuItemRepoForOrder()
	item := seedMenuItem(menuRepo, true)
	uc := newOrderUC(orderRepo, menuRepo)

	order, err := uc.CreateForTable(context.Background(), uuid.New(), []domain.OrderItemInput{
		{MenuItemID: item.ID, Quantity: 1},
	}, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	for _, next := range chain {
		order, err = uc.AdvanceStatus(context.Background(), order.ID, next)
		if err != nil {
			t.Fatalf("advance to %s: %v", next, err)
		}
		if order.Status != next {
			t.Errorf("expected status %s, got %s", next, order.Status)
		}
	}
}

func TestAdvanceStatus_invalidTransition(t *testing.T) {
	orderRepo := newMockOrderRepo()
	menuRepo := newMockMenuItemRepoForOrder()
	item := seedMenuItem(menuRepo, true)
	uc := newOrderUC(orderRepo, menuRepo)

	order, _ := uc.CreateForTable(context.Background(), uuid.New(), []domain.OrderItemInput{
		{MenuItemID: item.ID, Quantity: 1},
	}, "")

	// pending → ready is not a valid single step
	_, err := uc.AdvanceStatus(context.Background(), order.ID, domain.StatusReady)
	if !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestCancel_success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	menuRepo := newMockMenuItemRepoForOrder()
	item := seedMenuItem(menuRepo, true)
	uc := newOrderUC(orderRepo, menuRepo)

	order, _ := uc.CreateForTable(context.Background(), uuid.New(), []domain.OrderItemInput{
		{MenuItemID: item.ID, Quantity: 1},
	}, "")

	cancelled, err := uc.Cancel(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Errorf("expected cancelled, got %s", cancelled.Status)
	}
}

func TestCancel_alreadyCancelled(t *testing.T) {
	orderRepo := newMockOrderRepo()
	menuRepo := newMockMenuItemRepoForOrder()
	item := seedMenuItem(menuRepo, true)
	uc := newOrderUC(orderRepo, menuRepo)

	order, _ := uc.CreateForTable(context.Background(), uuid.New(), []domain.OrderItemInput{
		{MenuItemID: item.ID, Quantity: 1},
	}, "")
	_, _ = uc.Cancel(context.Background(), order.ID)

	_, err := uc.Cancel(context.Background(), order.ID)
	if !errors.Is(err, domain.ErrOrderCancelled) {
		t.Errorf("expected ErrOrderCancelled, got %v", err)
	}
}

func TestCancel_completedOrder(t *testing.T) {
	orderRepo := newMockOrderRepo()
	menuRepo := newMockMenuItemRepoForOrder()
	item := seedMenuItem(menuRepo, true)
	uc := newOrderUC(orderRepo, menuRepo)

	order, _ := uc.CreateForTable(context.Background(), uuid.New(), []domain.OrderItemInput{
		{MenuItemID: item.ID, Quantity: 1},
	}, "")
	for _, s := range []string{domain.StatusConfirmed, domain.StatusPreparing, domain.StatusReady, domain.StatusCompleted} {
		_, _ = uc.AdvanceStatus(context.Background(), order.ID, s)
	}

	_, err := uc.Cancel(context.Background(), order.ID)
	if !errors.Is(err, domain.ErrOrderCancelled) {
		t.Errorf("expected ErrOrderCancelled, got %v", err)
	}
}

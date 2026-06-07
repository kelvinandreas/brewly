# PRP — Orders, Payments, KDS (M3)

## Goal

Ship the core ordering loop end-to-end: customers place orders from their table (via table token), cashiers place orders at the counter, kitchen staff see live orders on a KDS board (SSE-powered), advance order status through the state machine, and cashiers record payments. When M3 ships, Brewly is a working point-of-sale system — orders flow, kitchen knows what to make, cashier knows when to collect.

## Context

- Architecture per `memory-bank/architecture.md`: handler → usecase → domain ← repository; SSE broker lives in `pkg/sse/`; errors wrap on crossings.
- Schema per `memory-bank/database-schema.md`: `orders`, `order_items`, `payments` tables already defined in `001_init.sql`. No new migration needed.
- `orders.status` state machine: `pending → confirmed → preparing → ready → completed`. Cancel is a separate side-channel (`POST /:id/cancel`) from any non-terminal state. Invalid transitions return `ErrInvalidStatusTransition`.
- `orders.source`: `customer_qr` (table-token caller) or `cashier_pos` (staff JWT caller).
- `order_items` captures `name_snapshot` and `price_minor_snapshot` at order time so menu renames/price changes never rewrite history.
- Payment is independent of status; it can be recorded once order is `ready` or `completed`.
- SSE: `EventSource` in the browser cannot send `Authorization` headers — accept access token as `?token=<jwt>` query param on the SSE endpoint. Per ADR-005 the backend flushes after every write and sends `: keep-alive\n\n` every 15 s. Handler closes on `r.Context().Done()`.
- All domain table rows soft-deleted; orders and payments are never soft-deleted in v1 (immutable audit trail).
- Roles: `owner` and `cashier` can place/cancel/advance orders and record payments. `kitchen` can advance status (confirming → preparing → ready) but cannot cancel or create.
- `GET /api/customer/orders/mine` — last 5 orders belonging to the table JWT's `tid` claim, ordered `created_at DESC`.

## File structure

### To create

**Backend**
- `backend/internal/domain/order.go` — `Order`, `OrderItem` structs + `OrderRepository` interface
- `backend/internal/domain/payment.go` — `Payment` struct + `PaymentRepository` interface
- `backend/internal/domain/errors.go` additions — `ErrOrderNotFound`, `ErrOrderCancelled`, `ErrInvalidStatusTransition`, `ErrPaymentConflict`
- `backend/internal/domain/constants.go` additions — order statuses, order sources, payment methods
- `backend/pkg/sse/broker.go` — generic fan-out SSE broker (`Broker`, `Event`, `Subscribe`, `Publish`, `NewBroker`)
- `backend/internal/repository/order_repo.go` — GORM impl of `OrderRepository`
- `backend/internal/repository/payment_repo.go` — GORM impl of `PaymentRepository`
- `backend/internal/usecase/order.go` — `OrderUsecase`: `CreateForTable`, `CreateForCashier`, `List`, `GetByID`, `AdvanceStatus`, `Cancel`
- `backend/internal/usecase/order_test.go` — table-driven tests for state machine transitions
- `backend/internal/usecase/payment.go` — `PaymentUsecase`: `Record`, `ListByOrder`
- `backend/internal/handler/order_handler.go` — staff order endpoints
- `backend/internal/handler/payment_handler.go` — payment endpoints
- `backend/internal/handler/sse_handler.go` — `SSEHandler`: `KitchenStream` (GET /api/sse/kitchen)
- `frontend/src/hooks/useOrders.ts` — staff order query + mutations
- `frontend/src/hooks/useCustomerOrder.ts` — customer POST order + GET mine
- `frontend/src/hooks/useKitchenSSE.ts` — `EventSource` subscription, returns live order list
- `frontend/src/routes/_auth.kitchen.tsx` — KDS board (kitchen/cashier/owner)
- `frontend/src/routes/_auth.cashier.tsx` — cashier POS: table picker → menu → cart → place order → payment
- `frontend/src/routes/_auth.orders.tsx` — owner/cashier order history list

### To modify

- `backend/internal/domain/errors.go` — append new sentinels
- `backend/internal/domain/constants.go` — append order/payment constants
- `backend/internal/handler/customer_handler.go` — add `PlaceOrder` (POST /api/customer/orders) and `MyOrders` (GET /api/customer/orders/mine)
- `backend/cmd/api/main.go` — wire order/payment/SSE repos, usecases, handlers, routes
- `frontend/src/types/api.ts` — add `Order`, `OrderItem`, `Payment`, `PlaceOrderRequest`, SSE event types
- `frontend/src/routes/table.$tableId.tsx` — add order placement UI (cart state, POST order, show my orders)
- `frontend/src/routes/_auth.dashboard.tsx` — add nav links to /kitchen, /cashier, /orders
- `frontend/src/main.tsx` — register kitchen, cashier, orders routes
- `memory-bank/progress.md` — tick M3 items

## Task breakdown

1. **Domain: order + payment entities** — `domain/order.go` and `domain/payment.go`. Define `Order`, `OrderItem`, `Payment` structs with GORM-compatible field names. Define `OrderRepository` and `PaymentRepository` interfaces. Commit: `feat(domain): add Order, OrderItem, Payment entities and repo interfaces`.

2. **Domain: errors + constants** — append to `domain/errors.go` and `domain/constants.go`. Errors: `ErrOrderNotFound`, `ErrOrderCancelled`, `ErrInvalidStatusTransition`, `ErrPaymentConflict`. Constants: `StatusPending/Confirmed/Preparing/Ready/Completed/Cancelled`, `SourceCustomerQR/CashierPOS`, `PaymentMethodCash/QRIS/Card`. Commit: `feat(domain): add order status, source, payment constants and errors`.

3. **SSE broker** — `pkg/sse/broker.go`. `Broker` holds `map[chan Event]struct{}` guarded by `sync.RWMutex`. `NewBroker()`, `Subscribe() (chan Event, func())` (func is unsubscribe), `Publish(Event)` fan-out with non-blocking send (drop if slow consumer). Commit: `feat(pkg): add generic SSE broker`.

4. **Order repository** — `repository/order_repo.go`. Methods: `Create(ctx, *Order) error` (inserts order + order_items in one transaction), `FindByID(ctx, uuid) (*Order, error)`, `List(ctx, filter OrderFilter) ([]Order, error)`, `UpdateStatus(ctx, id uuid, status string) error`, `ListByTable(ctx, tableID uuid, limit int) ([]Order, error)`. `OrderFilter`: `{Status *string, From *time.Time, To *time.Time}`. Preload `OrderItems` on FindByID and List. Commit: `feat(repo): add OrderRepo`.

5. **Payment repository** — `repository/payment_repo.go`. Methods: `Create(ctx, *Payment) error`, `ListByOrder(ctx, orderID uuid) ([]Payment, error)`. Commit: `feat(repo): add PaymentRepo`.

6. **Order usecase + tests** — `usecase/order.go`. Methods:
   - `CreateForTable(ctx, tableID uuid, items []OrderItemInput, note string) (*Order, error)` — validates items available, builds snapshots, calculates total, wraps Create in tx, publishes `order.created` SSE event.
   - `CreateForCashier(ctx, tableID, createdByUserID uuid, items []OrderItemInput, note string) (*Order, error)` — same logic, `source = cashier_pos`.
   - `List(ctx, filter OrderFilter) ([]Order, error)` — delegates to repo.
   - `GetByID(ctx, id uuid) (*Order, error)`.
   - `AdvanceStatus(ctx, id uuid, newStatus string) (*Order, error)` — validates transition via state machine, calls `repo.UpdateStatus`, publishes `order.status_changed`.
   - `Cancel(ctx, id uuid, reason string) (*Order, error)` — checks not already terminal, sets `cancelled`, publishes `order.cancelled`.
   
   Tests in `usecase/order_test.go`: create with unavailable item fails, valid create succeeds, invalid transition fails (pending→ready), valid transition chain, cancel cancelled order fails.
   Commit: `feat(usecase): add OrderUsecase with state machine and tests`.

7. **Payment usecase** — `usecase/payment.go`. `Record(ctx, orderID, recordedByUserID uuid, method string, amountMinor, receivedMinor int64) (*Payment, error)`. Validates method is valid, creates payment. `ListByOrder(ctx, orderID uuid)`. Commit: `feat(usecase): add PaymentUsecase`.

8. **SSE handler** — `handler/sse_handler.go`. `KitchenStream` GET `/api/sse/kitchen?token=<jwt>`: reads `token` query param → validates with `appMiddleware.VerifyAccessToken(secret)` helper → subscribes to kitchen broker → streams SSE frames until `ctx.Done()`. Also sends `15s` keep-alive comments. Commit: `feat(handler): add SSEHandler for kitchen stream`.

9. **Order handler** — `handler/order_handler.go`. Endpoints:
   - `GET /api/orders` — query params `status`, `from`, `to` (RFC3339).
   - `GET /api/orders/:id`
   - `POST /api/orders` (cashier/owner) — body `{tableId, items, note?}`
   - `PATCH /api/orders/:id/status` — body `{status}`
   - `POST /api/orders/:id/cancel` — body `{reason?}`
   Commit: `feat(handler): add OrderHandler`.

10. **Payment handler** — `handler/payment_handler.go`. `POST /api/orders/:id/payments`, `GET /api/orders/:id/payments`. Commit: `feat(handler): add PaymentHandler`.

11. **CustomerHandler additions** — extend `handler/customer_handler.go`. Add `PlaceOrder` (POST `/api/customer/orders`): reads tableID from ctx, validates items, calls `orderUC.CreateForTable`, returns `{order}`. Add `MyOrders` (GET `/api/customer/orders/mine`): calls `orderUC.ListByTable(ctx, tableID, 5)`, returns `{orders}`. Commit: `feat(handler): add customer order endpoints`.

12. **main.go wiring** — instantiate `sse.NewBroker()` (kitchen broker), `OrderRepo`, `PaymentRepo`, `OrderUsecase(orderRepo, menuItemRepo, kitchenBroker)`, `PaymentUsecase(paymentRepo, orderRepo)`, `OrderHandler`, `PaymentHandler`, `SSEHandler(kitchenBroker, authCfg.AccessSecret)`. Mount routes:
    - `/api/orders` (JWT cashier+owner for POST; JWT all roles for GET)
    - `/api/orders/:id` (JWT all roles)
    - `/api/orders/:id/status` (JWT all roles)
    - `/api/orders/:id/cancel` (JWT cashier+owner)
    - `/api/orders/:id/payments` (JWT cashier+owner)
    - `/api/sse/kitchen` (no middleware — handler validates token from query param)
    - `/api/customer/orders` and `/api/customer/orders/mine` under existing `RequireTableToken` group
    Commit: `feat(api): wire orders, payments, SSE routes in main.go`.

13. **`frontend/src/types/api.ts` additions** — `Order`, `OrderItem`, `Payment`, `PlaceOrderRequest`, `OrderItemInput`, `KitchenSSEEvent` (discriminated union: `{type: 'order.created'|'order.status_changed'|'order.cancelled', data: Order}`). Commit: `feat(frontend): add Order, Payment API types`.

14. **`frontend/src/hooks/useOrders.ts`** — query key `['orders', filters]`, `invalidateQueries` on mutations. Exports `useOrders(filters)` with `listQuery`, `createMutation` (cashier POS), `advanceStatusMutation`, `cancelMutation`. Commit: `feat(frontend): add useOrders hook`.

15. **`frontend/src/hooks/useCustomerOrder.ts`** — `placeOrderMutation` (POST /api/customer/orders using tableToken from `getTableToken()`), `myOrdersQuery` (GET /api/customer/orders/mine). Commit: `feat(frontend): add useCustomerOrder hook`.

16. **`frontend/src/hooks/useKitchenSSE.ts`** — `useKitchenSSE(accessToken)`: opens `EventSource('/api/sse/kitchen?token=…')`, listens on `message` event (parse JSON), maintains local `orders: Order[]` state via `useReducer` (add on `order.created`, update on `order.status_changed`, update on `order.cancelled`). Closes EventSource on unmount. Commit: `feat(frontend): add useKitchenSSE hook`.

17. **`frontend/src/routes/_auth.kitchen.tsx`** — KDS board. Uses `useKitchenSSE` + initial load from `useOrders({status: 'active'})`. Four columns: Pending, Confirmed, Preparing, Ready. Each card shows table label, items, time elapsed. Advance-status button per card. Roles: kitchen/cashier/owner. Commit: `feat(frontend): add kitchen KDS route`.

18. **`frontend/src/routes/_auth.cashier.tsx`** — Cashier POS. Three steps: (1) pick table from `useTables()`, (2) browse menu by category (`useCategories` + `useMenuItems`), add to cart, (3) review cart + submit via `createMutation`. After order placed, show order card with payment form (`PaymentUsecase.Record`). Commit: `feat(frontend): add cashier POS route`.

19. **`frontend/src/routes/_auth.orders.tsx`** — Order history list. `useOrders()` with status filter tabs (all / active / completed / cancelled). Each row shows table, status badge, total, created time. Commit: `feat(frontend): add orders history route`.

20. **`frontend/src/routes/table.$tableId.tsx` update** — add cart state (`items: {menuItem, qty}[]`) below menu grid. "Add to cart" button on each item card. Cart drawer/panel shows items + total + Place Order button → `placeOrderMutation`. After success show order status. Commit: `feat(frontend): add order placement to customer table route`.

21. **Wire routes + dashboard** — update `main.tsx` (add kitchenRoute, cashierRoute, ordersRoute to authTree), update `_auth.dashboard.tsx` (add nav links to /kitchen, /cashier, /orders). Commit: `feat(frontend): wire kitchen, cashier, orders routes`.

22. **Memory bank** — tick M3 items in `progress.md`. Commit: `chore(docs): update memory-bank after M3 orders, payments, KDS`.

## Pseudocode (CreateForTable — trickiest function)

```go
// usecase/order.go
type OrderItemInput struct {
    MenuItemID uuid.UUID
    Quantity   int
}

func (u *OrderUsecase) createOrder(
    ctx context.Context,
    tableID uuid.UUID,
    createdByUserID *uuid.UUID, // nil for customer_qr
    source string,
    inputs []OrderItemInput,
    note string,
) (*domain.Order, error) {
    if len(inputs) == 0 {
        return nil, fmt.Errorf("usecase.createOrder: at least one item required")
    }

    var items []domain.OrderItem
    var totalMinor int64

    for _, inp := range inputs {
        mi, err := u.menuItemRepo.FindByID(ctx, inp.MenuItemID)
        if err != nil {
            return nil, fmt.Errorf("usecase.createOrder: %w", err) // ErrMenuItemNotFound propagates
        }
        if !mi.IsAvailable {
            return nil, domain.ErrMenuItemUnavailable
        }
        items = append(items, domain.OrderItem{
            MenuItemID:          mi.ID,
            NameSnapshot:        mi.Name,
            PriceMinorSnapshot:  mi.PriceMinor,
            Quantity:            inp.Quantity,
        })
        totalMinor += mi.PriceMinor * int64(inp.Quantity)
    }

    order := &domain.Order{
        TableID:           tableID,
        Status:            domain.StatusPending,
        Source:            source,
        TotalMinor:        totalMinor,
        Note:              note,
        CreatedByUserID:   createdByUserID,
        Items:             items,
    }

    if err := u.orderRepo.Create(ctx, order); err != nil {
        return nil, fmt.Errorf("usecase.createOrder: %w", err)
    }

    // Publish after DB commit — failures here don't roll back the order.
    payload, _ := json.Marshal(order)
    u.kitchenBroker.Publish(sse.Event{Type: "order.created", Payload: payload})

    return order, nil
}
```

```go
// State machine: valid next states per current status
var validTransitions = map[string]string{
    domain.StatusPending:   domain.StatusConfirmed,
    domain.StatusConfirmed: domain.StatusPreparing,
    domain.StatusPreparing: domain.StatusReady,
    domain.StatusReady:     domain.StatusCompleted,
}

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
    payload, _ := json.Marshal(order)
    u.kitchenBroker.Publish(sse.Event{Type: "order.status_changed", Payload: payload})
    return order, nil
}
```

## Validation plan

- `make lint` clean (golangci-lint + eslint + prettier --check).
- `make test` — `usecase/order_test.go` must cover: create with unavailable item fails, create success calculates correct total, invalid state transition (pending→ready) returns `ErrInvalidStatusTransition`, valid chain (pending→confirmed→preparing→ready→completed) succeeds, cancel on completed order returns error.
- Manual smoke (with `make dev`):
  1. Create categories + items in `/menu` (owner).
  2. Create a table in `/tables` (owner), scan QR.
  3. Customer adds items to cart and places order — order appears with `pending` status in `/kitchen`.
  4. Kitchen advances: confirmed → preparing → ready.
  5. Cashier opens `/cashier`, records payment for the ready order.
  6. Order moves to `completed`.
  7. Owner opens `/orders`, sees completed order with correct total.
  8. `curl -N "http://localhost:8080/api/sse/kitchen?token=<access_token>"` streams keep-alive lines.

## Out of scope

- Partial payments / split bills.
- Order editing after placement (cancel + re-create is the v1 flow).
- Customer-side real-time order status polling (customer can refresh `/table/$tableId` to see status from `GET /api/customer/orders/mine`).
- Refunds.
- Printer/receipt integration.

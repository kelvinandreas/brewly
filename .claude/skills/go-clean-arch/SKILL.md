---
name: go-clean-arch
description: Enforce Brewly's clean-architecture layer rules when editing Go files in handler/, usecase/, repository/, or domain/. Triggers on dependency injection, business logic placement, and error wrapping in backend Go code.
---

Brewly's backend is divided into four layers with one direction of dependency. Cross at your peril.

```
handler  ─►  usecase  ─►  domain  ◄─  repository
```

## Rules

1. **`domain/`** — Entities, repository interfaces, sentinel errors, constants. Allowed imports: stdlib + `github.com/google/uuid`. Nothing else. If you find yourself importing GORM, Chi, or JWT here, stop.

2. **`usecase/`** — Business logic. Depends on `domain.*Repository` interfaces only. **No** `*http.Request`, no `gorm.DB`, no JSON tags. Returns `domain` types or sentinel errors.

3. **`repository/`** — GORM lives here only. Implement `domain.*Repository`. Convert GORM `ErrRecordNotFound` into `domain.ErrXxxNotFound`.

4. **`handler/`** — Decode JSON DTO → call usecase → encode via `pkg/response`. Maps domain sentinel errors → HTTP statuses. No business logic — if you find an `if` with business meaning here, move it.

5. **Errors** — Wrap on every layer crossing: `fmt.Errorf("layer.Method: %w", err)`. Caller uses `errors.Is(err, domain.ErrXxx)`.

## Wiring (in `cmd/api/main.go`)

```
db := postgres.Open(...)
userRepo := repository.NewUserRepo(db)
authUC   := usecase.NewAuth(userRepo, cfg)
authH    := handler.NewAuth(authUC)
router.Mount("/api/auth", authH.Routes())
```

Dependency flow is always: low-level → high-level.

## DO

```go
// usecase/order.go
func (u *OrderUsecase) Create(ctx context.Context, in domain.NewOrder) (domain.Order, error) {
    if len(in.Items) == 0 {
        return domain.Order{}, domain.ErrEmptyOrder
    }
    order, err := u.repo.Insert(ctx, in)
    if err != nil {
        return domain.Order{}, fmt.Errorf("usecase.Order.Create: %w", err)
    }
    return order, nil
}
```

```go
// handler/order_handler.go
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    var dto createOrderDTO
    if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid_body", "could not parse body")
        return
    }
    order, err := h.uc.Create(r.Context(), dto.toDomain())
    switch {
    case errors.Is(err, domain.ErrEmptyOrder):
        response.Error(w, http.StatusUnprocessableEntity, "empty_order", "order must have items")
    case err != nil:
        response.Error(w, http.StatusInternalServerError, "internal", "unexpected error")
    default:
        response.OK(w, order)
    }
}
```

## DON'T

```go
// handler/order_handler.go — WRONG: business validation in handler
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    var dto createOrderDTO
    json.NewDecoder(r.Body).Decode(&dto)
    if len(dto.Items) == 0 {              // ← belongs in usecase
        http.Error(w, "empty", 400)
        return
    }
    var total int64                        // ← business math in handler
    for _, it := range dto.Items {
        total += it.Price * int64(it.Qty)
    }
    h.db.Create(&Order{Total: total})      // ← raw GORM in handler!
}
```

## Common slips
- Importing `gorm.io/gorm` outside `repository/` — even a `*gorm.DB` parameter type leaks the dependency.
- Returning `*gorm.DB.Error` from a usecase — wrap or convert to a sentinel first.
- Putting validation tags (`validate:"required"`) on domain structs — they belong on handler DTOs.

# Architecture

## System diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Customer phone                          │
│                                                                 │
│   Scan QR  ─►  https://cafe.example.com/table/3?token=<jwt>     │
│                                                                 │
└──────────────────────────────┬──────────────────────────────────┘
                               │ Bearer table-token
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Frontend (React 19 + Vite SPA)                 │
│                                                                 │
│  routes/table/$tableId  ──┐                                     │
│  routes/_auth/dashboard  ─┼─► hooks/use* ──► lib/api.ts ────┐   │
│  routes/kitchen          ─┘                                 │   │
└─────────────────────────────────────────────────────────────┼───┘
                                                              │
                                                              ▼ HTTP/JSON + SSE
┌─────────────────────────────────────────────────────────────────┐
│                       Backend (Go + Chi)                        │
│                                                                 │
│   middleware ─► handler ─► usecase ─► repository (GORM)         │
│        │             │          │                               │
│        │             │          └─► domain interfaces           │
│        │             └─► pkg/response, pkg/jwt                  │
│        └─► jwt_auth, table_token, rate_limiter, sse-keep-alive  │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
                ┌────────────────────────────┐
                │      PostgreSQL 16         │
                │  users / tables /          │
                │  categories / menu_items / │
                │  orders / order_items /    │
                │  payments / song_requests  │
                └────────────────────────────┘
```

## Component responsibilities

- **`cmd/api/main.go`** — wires config → DB → repositories → usecases → handlers → Chi router, starts HTTP server with graceful shutdown.
- **`internal/middleware/`** — request-scoped concerns: JWT validation (staff), table-token validation (customer), rate limit lookup, structured logging, panic recovery, CORS.
- **`internal/handler/`** — HTTP boundary. Decodes request JSON, calls a usecase, encodes response via `pkg/response`. Owns nothing.
- **`internal/usecase/`** — business logic. Transactions, validation, orchestration. Depends on `domain.*Repository` interfaces.
- **`internal/repository/`** — persistence. GORM lives here only. One repo per aggregate.
- **`internal/domain/`** — types + interfaces + sentinel errors + constants. No imports outside std + uuid.
- **`pkg/jwt`** — sign + verify for staff JWT and table-token JWT (different secrets, different claims).
- **`pkg/youtube`** — wraps Data API v3 search endpoint. Returns lean DTOs.

## Data flow — customer places order

1. Browser hits `/table/3?token=…` → React extracts token from URL → strips it from history.replaceState → stores in `lib/auth.ts` memory store.
2. `MenuGrid` calls `useMenu()` which calls `GET /api/customer/menu` with `Authorization: Bearer <table-token>`.
3. Customer taps "Order" → `useCreateOrder()` → `POST /api/customer/orders` with token + items.
4. Backend `table_token` middleware verifies signature, checks `tables.token_version` matches claim — rejects if regenerated.
5. `orderUsecase.CreateForTable(ctx, tableID, items)` validates items exist and are available, opens a transaction, inserts `orders` row + `order_items`, returns `Order`.
6. SSE endpoint `/api/kitchen/stream` pushes a `order.created` event — KDS board mounts a new card without polling.

## Data flow — owner regenerates QR

1. Owner clicks "Regenerate QR" on a table row → `useRegenerateTable(tableID)` → `POST /api/tables/:id/regenerate-token`.
2. `tableUsecase.RegenerateToken` increments `tables.token_version`, returns the new signed token + QR PNG bytes.
3. Old token signatures still verify cryptographically, but the `tvr` claim no longer matches `tables.token_version` → middleware rejects.

## Key design decisions (anchor links to ADRs)

- Why Chi over Gin → [ADR-001](../docs/ADR/ADR-001-chi-over-gin.md)
- Why Clean Architecture → [ADR-002](../docs/ADR/ADR-002-clean-architecture.md)
- Why GORM with raw SQL fallback → [ADR-003](../docs/ADR/ADR-003-gorm-with-raw-fallback.md)
- Why JWT `token_version` over a blacklist → [ADR-004](../docs/ADR/ADR-004-table-token-security.md)
- Why SSE over WebSocket → [ADR-005](../docs/ADR/ADR-005-sse-for-realtime.md)
- Why YouTube Data API v3 with API key → [ADR-006](../docs/ADR/ADR-006-youtube-api-for-song-search.md)

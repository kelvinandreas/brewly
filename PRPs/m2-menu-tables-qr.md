# PRP — M2 Menu, Tables & QR

## Goal

Ship menu management (categories + items), table management with QR code generation, the table-token JWT middleware, and the customer-facing menu browse endpoint. After this milestone: an owner can build a full menu, create numbered tables, and hand a customer a QR code that opens a read-only menu view in their browser. This is the prerequisite for M3 (ordering).

## Context

- Architecture (`memory-bank/architecture.md`): clean layers unchanged — handler → usecase → domain ← repository. No GORM outside `repository/`. Table token middleware lives in `internal/middleware/table_token.go`.
- Schema (`memory-bank/database-schema.md`): `categories`, `menu_items`, `tables` all created in `001_init.sql`. No new migrations needed in M2.
- API contracts (`memory-bank/api-contracts.md`): §Categories, §Menu items, §Tables (including `POST /regenerate-token` and `GET /qr.png`), and §Customer `/api/customer/menu` are all specified; M2 implements them.
- Table-token design (`docs/ADR/ADR-004-table-token-security.md`):
  - Claims: `{tid: tableUUID, tvr: int, jti: uuid, iat, exp}`. TTL 4h. Signed with `TABLE_TOKEN_SECRET`.
  - Middleware: verify sig → check exp → lookup `tables.token_version` → compare `tvr` → inject `tableID` + `tokenJTI` into context.
  - `POST /api/tables/:id/regenerate-token`: bumps `token_version`, re-signs, returns new token + QR.
- QR URL shape: `{FRONTEND_URL}/table/{tableID}?token={signedJWT}`. `FRONTEND_URL` read from env (default `http://localhost:5173`).
- `pkg/tabletoken` is a *separate* package from `pkg/jwt` — different claims struct, different validation rules (`tvr` check, no `role` claim).
- QR library: `github.com/skip2/go-qrcode` — single call `qrcode.Encode(content, qrcode.Medium, 256)` returns PNG bytes.
- Customer menu response shape: `{ categories: [ { id, name, displayOrder, items: [ {id, name, description, priceMinor, imageUrl, isAvailable} ] } ] }` — only `isAvailable=true` items, categories ordered by `display_order ASC`.
- Frontend uses `createRoute` (manual routing), not `createFileRoute` — established in M1.
- `src/lib/auth.ts` already has the module-scope token store; the customer table-token will use a **separate** `src/lib/tableAuth.ts` store (same pattern, different variable) so the two token types never collide.

## File structure

### To create — backend

- `backend/internal/domain/category.go` — `Category` struct + `CategoryRepository` interface
- `backend/internal/domain/menu_item.go` — `MenuItem` struct + `MenuItemRepository` interface
- `backend/internal/domain/table.go` — `Table` struct + `TableRepository` interface
- `backend/pkg/tabletoken/tabletoken.go` — `Sign`, `Verify`, `TableClaims`
- `backend/pkg/tabletoken/tabletoken_test.go` — verify, expired, tampered, tvr-mismatch path
- `backend/pkg/qrcode/qrcode.go` — `Generate(content string) ([]byte, error)`
- `backend/internal/repository/category_repo.go`
- `backend/internal/repository/menu_item_repo.go`
- `backend/internal/repository/table_repo.go` — includes `IncrementTokenVersion`
- `backend/internal/middleware/table_token.go` — `RequireTableToken(secret string, tableRepo domain.TableRepository) func(http.Handler) http.Handler`
- `backend/internal/usecase/category.go` — `CategoryUsecase`: `List`, `Create`, `Update`, `SoftDelete`
- `backend/internal/usecase/menu_item.go` — `MenuItemUsecase`: `List`, `Create`, `Update`, `SoftDelete`
- `backend/internal/usecase/table.go` — `TableUsecase`: `List`, `Create`, `Update`, `SoftDelete`, `RegenerateToken`, `GetQRPNG`
- `backend/internal/usecase/table_test.go` — Create signs a valid token, RegenerateToken bumps version + invalidates old token
- `backend/internal/handler/category_handler.go`
- `backend/internal/handler/menu_item_handler.go`
- `backend/internal/handler/table_handler.go` — includes `GetQR` that serves `image/png`
- `backend/internal/handler/customer_handler.go` — `GET /api/customer/menu`

### To create — frontend

- `frontend/src/lib/tableAuth.ts` — module-scope `_tableToken: string | null`; `setTableToken`, `getTableToken`, `clearTableToken`
- `frontend/src/hooks/useCategories.ts` — query key `['categories']`; mutations `createCategory`, `updateCategory`, `deleteCategory`
- `frontend/src/hooks/useMenuItems.ts` — query key `['menu-items', {categoryId?, availableOnly?}]`; mutations `createMenuItem`, `updateMenuItem`, `deleteMenuItem`
- `frontend/src/hooks/useTables.ts` — query key `['tables']`; mutations `createTable`, `updateTable`, `deleteTable`, `regenerateToken`
- `frontend/src/routes/_auth.menu.tsx` — owner menu management page
- `frontend/src/routes/_auth.tables.tsx` — owner tables page with QR modal
- `frontend/src/routes/table.$tableId.tsx` — customer QR landing page

### To modify

- `backend/internal/domain/errors.go` — add `ErrCategoryNotFound`, `ErrMenuItemNotFound`, `ErrTableNotFound`, `ErrTableLabelTaken`, `ErrMenuItemUnavailable`
- `backend/internal/domain/constants.go` — add `ContextKeyTableID`, `ContextKeyTokenJTI`
- `backend/go.mod` / `go.sum` — after `go get github.com/skip2/go-qrcode`
- `backend/cmd/api/main.go` — wire 3 new repo/usecase/handler sets + table-token middleware + customer route
- `frontend/src/types/api.ts` — add `Category`, `MenuItem`, `Table`, `CustomerMenuCategory`, `CustomerMenuResponse`
- `frontend/src/routes/_auth.dashboard.tsx` — add nav links to `/menu` and `/tables`
- `frontend/src/main.tsx` — add `menuRoute`, `tablesRoute`, `tableIdRoute` to route tree
- `memory-bank/api-contracts.md` — mark §Categories, §Menu items, §Tables, §Customer menu as implemented
- `memory-bank/progress.md` — tick M2 items

## Task breakdown

Each numbered step = one atomic commit.

1. **Go dep** — `cd backend && go get github.com/skip2/go-qrcode`. Commit: `chore(deps): add go-qrcode`.

2. **Domain additions** — `domain/category.go`, `domain/menu_item.go`, `domain/table.go`. Append `ErrCategoryNotFound`, `ErrMenuItemNotFound`, `ErrTableNotFound`, `ErrTableLabelTaken`, `ErrMenuItemUnavailable` to `domain/errors.go`. Append `ContextKeyTableID`, `ContextKeyTokenJTI` to `domain/constants.go`. `go build ./internal/domain/...`. Commit: `feat(domain): add Category, MenuItem, Table entities and errors`.

3. **`pkg/tabletoken` + tests** — `tabletoken.go` + `tabletoken_test.go`. `go test ./pkg/tabletoken/...`. Commit: `feat(pkg): add tabletoken Sign/Verify with tests`.

4. **`pkg/qrcode`** — `qrcode.go` wraps `github.com/skip2/go-qrcode`. `go build ./pkg/qrcode/...`. Commit: `feat(pkg): add qrcode.Generate helper`.

5. **`repository/category_repo.go`** — `go build ./internal/repository/...`. Commit: `feat(repo): add CategoryRepo`.

6. **`repository/menu_item_repo.go`** — Commit: `feat(repo): add MenuItemRepo`.

7. **`repository/table_repo.go`** — includes `IncrementTokenVersion(ctx, id) error`. Commit: `feat(repo): add TableRepo`.

8. **`middleware/table_token.go`** — `go build ./internal/middleware/...`. Commit: `feat(middleware): add RequireTableToken middleware`.

9. **`usecase/category.go`** — Commit: `feat(usecase): add CategoryUsecase`.

10. **`usecase/menu_item.go`** — Commit: `feat(usecase): add MenuItemUsecase`.

11. **`usecase/table.go` + tests** — `TableUsecase` with `Create`, `Update`, `SoftDelete`, `RegenerateToken`, `GetQRPNG`. Tests: Create returns valid token; RegenerateToken bumps version and old-token verify returns error. `go test ./internal/usecase/...`. Commit: `feat(usecase): add TableUsecase with tests`.

12. **`handler/category_handler.go`** — 4 thin handlers. Commit: `feat(handler): add CategoryHandler`.

13. **`handler/menu_item_handler.go`** — 4 thin handlers. Commit: `feat(handler): add MenuItemHandler`.

14. **`handler/table_handler.go`** — 6 handlers. `GetQR` writes `Content-Type: image/png` directly. Commit: `feat(handler): add TableHandler including QR endpoint`.

15. **`handler/customer_handler.go`** — `GET /api/customer/menu`. Commit: `feat(handler): add CustomerHandler — GET /api/customer/menu`.

16. **Wire `main.go`** — add new repos/usecases/handlers; mount `/api/categories`, `/api/menu-items`, `/api/tables` under `RequireAuth(owner)`; mount `/api/customer/menu` under `RequireTableToken`. `go build ./...`. Smoke: `curl /healthz` still 200. Commit: `feat(api): wire menu, tables, customer routes in main.go`.

17. **`frontend/src/types/api.ts` additions** — Commit: `feat(frontend): add Category, MenuItem, Table API types`.

18. **`frontend/src/lib/tableAuth.ts`** — Commit: `feat(frontend): add table token memory store`.

19. **`frontend/src/hooks/useCategories.ts`** — Commit: `feat(frontend): add useCategories hook`.

20. **`frontend/src/hooks/useMenuItems.ts`** — Commit: `feat(frontend): add useMenuItems hook`.

21. **`frontend/src/hooks/useTables.ts`** — Commit: `feat(frontend): add useTables hook`.

22. **`frontend/src/routes/_auth.menu.tsx`** — owner menu management page. Commit: `feat(frontend): add /menu owner page`.

23. **`frontend/src/routes/_auth.tables.tsx`** — owner tables page with QR modal. Commit: `feat(frontend): add /tables owner page`.

24. **`frontend/src/routes/table.$tableId.tsx`** — customer landing page. Commit: `feat(frontend): add /table/:tableId customer route`.

25. **Wire frontend + update dashboard** — update `_auth.dashboard.tsx` with nav links; add all new routes to `main.tsx` route tree. `tsc --noEmit` + `pnpm lint` clean. Commit: `feat(frontend): wire menu/tables/customer routes, update dashboard nav`.

26. **Memory bank** — tick M2 items in `progress.md`; no schema drift (schema was complete from M1); update `api-contracts.md` to mark endpoints implemented. Commit: `docs(memory-bank): mark M2 items complete`.

## Pseudocode — TableUsecase.Create and RegenerateToken (trickiest)

```
// internal/usecase/table.go

type TableConfig struct {
    TableTokenSecret string
    FrontendURL      string
}

type TableUsecase struct {
    repo      domain.TableRepository
    cfg       TableConfig
}

func (u *TableUsecase) Create(ctx context.Context, label string) (*domain.Table, string, string, error) {
    // label → qrToken, qrURL
    table := &domain.Table{ID: uuid.New(), Label: label, TokenVersion: 1}
    if err := u.repo.Create(ctx, table); err != nil {
        // ErrTableLabelTaken bubbles up from repo
        return nil, "", "", fmt.Errorf("usecase.Table.Create: %w", err)
    }
    token := tabletoken.Sign(table.ID, table.TokenVersion, u.cfg.TableTokenSecret)
    qrURL := fmt.Sprintf("%s/table/%s?token=%s", u.cfg.FrontendURL, table.ID, token)
    return table, token, qrURL, nil
}

func (u *TableUsecase) RegenerateToken(ctx context.Context, tableID uuid.UUID) (string, string, error) {
    if err := u.repo.IncrementTokenVersion(ctx, tableID); err != nil {
        return "", "", fmt.Errorf("usecase.Table.RegenerateToken: %w", err)
    }
    table, err := u.repo.FindByID(ctx, tableID)
    if err != nil {
        return "", "", fmt.Errorf("usecase.Table.RegenerateToken find: %w", err)
    }
    token := tabletoken.Sign(table.ID, table.TokenVersion, u.cfg.TableTokenSecret)
    qrURL := fmt.Sprintf("%s/table/%s?token=%s", u.cfg.FrontendURL, table.ID, token)
    return token, qrURL, nil
}
```

```
// pkg/tabletoken/tabletoken.go

type TableClaims struct {
    TableID      string `json:"tid"`
    TokenVersion int    `json:"tvr"`
    JTI          string `json:"jti"`
}

type rawTableClaims struct {
    jwt.RegisteredClaims
    TableID      string `json:"tid"`
    TokenVersion int    `json:"tvr"`
    JTI          string `json:"jti"`
}

func Sign(tableID uuid.UUID, tokenVersion int, secret string) string {
    now := time.Now()
    raw := rawTableClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(4 * time.Hour)),
        },
        TableID:      tableID.String(),
        TokenVersion: tokenVersion,
        JTI:          uuid.New().String(),
    }
    tok := jwt.NewWithClaims(jwt.SigningMethodHS256, raw)
    signed, _ := tok.SignedString([]byte(secret))
    return signed
}

func Verify(tokenStr, secret string) (*TableClaims, error) {
    // ... same pattern as pkg/jwt.Verify
    // Returns TableClaims{TableID, TokenVersion, JTI} on success
    // Returns ErrExpired or ErrInvalid on failure
}
```

```
// internal/middleware/table_token.go

func RequireTableToken(secret string, tableRepo domain.TableRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            header := r.Header.Get("Authorization")
            if !strings.HasPrefix(header, "Bearer ") {
                response.Error(w, 401, "unauthorized", "missing table token")
                return
            }
            claims, err := tabletoken.Verify(strings.TrimPrefix(header, "Bearer "), secret)
            if err != nil {
                response.Error(w, 401, "unauthorized", "invalid table token")
                return
            }
            tableID, _ := uuid.Parse(claims.TableID)
            table, err := tableRepo.FindByID(r.Context(), tableID)
            if err != nil || table.DeletedAt != nil {
                response.Error(w, 401, "unauthorized", "table not found")
                return
            }
            if table.TokenVersion != claims.TokenVersion {
                response.Error(w, 401, "token_version_mismatch", "QR code has been regenerated")
                return
            }
            ctx := context.WithValue(r.Context(), domain.ContextKeyTableID, tableID.String())
            ctx = context.WithValue(ctx, domain.ContextKeyTokenJTI, claims.JTI)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Validation plan

```bash
# After each commit:
cd backend && go build ./... && go vet ./...

# After step 3 (tabletoken):
go test ./pkg/tabletoken/... -v

# After step 11 (table usecase):
go test ./internal/usecase/... -v

# After step 16 (main.go wired) — with make dev running:
curl -s localhost:8080/healthz | jq .
# → {"status":"ok"}

# Register + login to get ACCESS token
ACCESS=$(curl -s -X POST localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' -c /tmp/brew.jar \
  -d '{"email":"owner@brew.ly","password":"secret123"}' | jq -r .data.accessToken)

# Categories
curl -s -X POST localhost:8080/api/categories \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"name":"Coffee","displayOrder":1}' | jq .
# → {success:true, data:{category}}

CAT_ID=$(curl -s localhost:8080/api/categories \
  -H "Authorization: Bearer $ACCESS" | jq -r '.data.categories[0].id')

# Menu items
curl -s -X POST localhost:8080/api/menu-items \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d "{\"categoryId\":\"$CAT_ID\",\"name\":\"Espresso\",\"priceMinor\":18000}" | jq .

# Tables
curl -s -X POST localhost:8080/api/tables \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"label":"1"}' | jq .
# → {success:true, data:{table, qrToken, qrUrl}}

TABLE_ID=$(curl -s localhost:8080/api/tables \
  -H "Authorization: Bearer $ACCESS" | jq -r '.data.tables[0].id')
TABLE_TOKEN=$(curl -s -X POST localhost:8080/api/tables \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"label":"2"}' | jq -r '.data.qrToken')

# Customer menu — using table token
curl -s localhost:8080/api/customer/menu \
  -H "Authorization: Bearer $TABLE_TOKEN" | jq .
# → {success:true, data:{categories:[{...items:[...]}]}}

# QR PNG
curl -s localhost:8080/api/tables/$TABLE_ID/qr.png \
  -H "Authorization: Bearer $ACCESS" -o /tmp/qr.png && file /tmp/qr.png
# → PNG image data, 256 x 256

# Regenerate token — old TABLE_TOKEN should now 401
curl -s -X POST localhost:8080/api/tables/$TABLE_ID/regenerate-token \
  -H "Authorization: Bearer $ACCESS" | jq .
curl -s localhost:8080/api/customer/menu \
  -H "Authorization: Bearer $TABLE_TOKEN" | jq .error
# → "unauthorized"

# After step 25 (frontend wired):
# Open http://localhost:5173/login → login as owner
# Navigate to /menu → create category → create item → toggle availability
# Navigate to /tables → create table → click QR button → modal shows QR image
# Click regenerate → QR changes
# Open http://localhost:5173/table/<id>?token=<token> → menu grid loads
```

## Out of scope

- Image upload for menu items (M2 accepts `imageUrl` as a plain string — owner pastes a URL; file upload deferred post-v1).
- Customer order placement (`POST /api/customer/orders`) — M3.
- Table-token rate limiting for song requests — M4.
- Cashier-facing menu visibility toggles beyond `isAvailable` — M2 ships the owner toggle.
- `GET /api/tables/:id/qr.png` re-generation on every call — the endpoint signs a fresh token each time using the current `token_version` (no caching needed at cafe scale).

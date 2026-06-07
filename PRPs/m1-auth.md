# PRP — M1 Auth

## Goal

Ship staff authentication end-to-end so that Brewly can be safely deployed for the first time. An empty database lets the first visitor self-register as `owner`; subsequent staff log in to receive a 15-minute access JWT (memory only) plus a 7-day httpOnly refresh cookie; the backend middleware gates every protected route by role. This PRP also commits the full initial database schema (`001_init.sql`) — all M2–M4 tables are defined here so they never need to be touched again in later milestones. After this milestone every other route has a working auth foundation to build on.

## Context

- Architecture (`memory-bank/architecture.md`): clean layers — `handler → usecase → domain ← repository`. Errors wrap at every layer crossing.
- `users` table per `memory-bank/database-schema.md`: `id uuid PK`, `email text UNIQUE NOT NULL`, `password_hash text NOT NULL` (bcrypt 12), `name`, `role CHECK(owner|cashier|kitchen)`, `last_refresh_jti text NULL`, soft delete.
- Auth endpoints per `memory-bank/api-contracts.md` §Auth: 5 endpoints — `register-owner`, `login`, `refresh`, `logout`, `me`.
- Access token: HS256, 15 min, claims `{sub, role, jti, iat, exp}`, signed with env `JWT_SECRET`.
- Refresh token: HS256, 7 days, claims `{sub, jti, iat, exp}`, signed with `REFRESH_SECRET`. Cookie: `name=BrewlyRefresh; Path=/api/auth; HttpOnly; Secure (prod); SameSite=Lax`.
- Refresh rotation: each `/refresh` issues a new refresh token and stores its JTI in `users.last_refresh_jti`. If incoming JTI ≠ stored JTI → replay, return 401. This invalidates stolen tokens.
- Frontend never touches `localStorage`/`sessionStorage` for tokens. Access token lives in a module-scope variable in `src/lib/auth.ts`; on app boot the SPA silently calls `POST /api/auth/refresh` (cookie auto-attached) to hydrate it.
- ADR-002 (clean arch): zero business logic in handlers. ADR-001 (Chi): Chi v5 router.
- TanStack Router v1 file-based routing already bootstrapped (`main.tsx`, `__root.tsx`).
- Frontend is missing `@hookform/resolvers` — must install before wiring the Zod resolver.
- Backend `go.mod` is missing: `golang-jwt/jwt/v5`, `golang.org/x/crypto`, `gorm.io/gorm`, `gorm.io/driver/postgres`, `github.com/google/uuid`, `github.com/go-playground/validator/v10`.

## File structure

### To create — backend

- `backend/migrations/001_init.sql` — full schema: `set_updated_at()` trigger fn, all 8 tables (users, tables, categories, menu_items, orders, order_items, payments, song_requests), all indexes
- `backend/migrations/002_seed_owner.sql` — comment-only placeholder; registration is self-serve via API
- `backend/internal/domain/user.go` — `User` struct + `UserRepository` interface
- `backend/internal/domain/errors.go` — `ErrUserNotFound`, `ErrEmailTaken`, `ErrInvalidCredentials`, `ErrOwnerExists`, `ErrUnauthorized`, `ErrForbidden`
- `backend/internal/domain/constants.go` — role strings, cookie name, `ContextKeyUserID`, `ContextKeyRole`
- `backend/internal/repository/user_repo.go` — GORM; methods: `Create`, `FindByEmail`, `FindByID`, `ExistsByRole`, `SetLastRefreshJTI`, `Update`, `SoftDelete`, `List`
- `backend/internal/usecase/auth.go` — `AuthUsecase`: `RegisterOwner`, `Login`, `Refresh`, `Logout`, `Me`
- `backend/internal/usecase/auth_test.go` — table-driven; covers register-twice, wrong password, refresh rotation replay, expired refresh
- `backend/internal/usecase/user.go` — `UserUsecase`: `List`, `Create`, `Update`, `SoftDelete` (owner-gated at handler level)
- `backend/internal/handler/auth_handler.go` — thin; parses JSON, calls usecase, sets/clears cookie via helpers
- `backend/internal/handler/user_handler.go` — owner-only CRUD
- `backend/internal/middleware/jwt_auth.go` — `RequireAuth(roles ...string) func(http.Handler) http.Handler`
- `backend/pkg/jwt/jwt.go` — `Sign(claims JWTClaims, secret string, ttl time.Duration) string`, `Verify(token, secret string) (*JWTClaims, error)`
- `backend/pkg/jwt/jwt_test.go` — tampered token, expired, valid, wrong secret
- `backend/pkg/response/response.go` — `OK(w, data)`, `OKMsg(w, data, msg)`, `Error(w, status, code, msg)`, `ValidationError(w, details)`
- `backend/pkg/db/db.go` — `Open(dsn string) (*gorm.DB, error)` wraps gorm postgres open + ping

### To create — frontend

- `frontend/src/types/api.ts` — `User`, `ApiResponse<T>`, `ApiError`, `LoginRequest`, `RegisterOwnerRequest`, `TokenResponse`
- `frontend/src/lib/auth.ts` — module-scope `_accessToken: string | null`; exports `setAccessToken`, `getAccessToken`, `clearAccessToken`, `refreshAccess(): Promise<string | null>`
- `frontend/src/lib/api.ts` — `apiFetch(path, init)` — injects `Authorization: Bearer` header from `getAccessToken()`, handles 401 by calling `refreshAccess()` and retrying once, throws `ApiError` on non-2xx
- `frontend/src/lib/currency.ts` — `formatIDR(minor: number): string`
- `frontend/src/hooks/useAuth.ts` — query key `['auth', 'me']`; queries `/api/auth/me`; mutations `login`, `logout`, `registerOwner`; on `login` success: `setAccessToken(data.accessToken)` then invalidate `['auth','me']`; on `logout`: `clearAccessToken()` then navigate to `/login`
- `frontend/src/routes/login.tsx` — `createFileRoute('/login')`. Shows login form by default; if `GET /api/auth/me` returns `ownerNotExists` code, toggles to register-owner form. Both use RHF + Zod.
- `frontend/src/routes/_auth.tsx` — `createFileRoute('/_auth')`. `beforeLoad`: calls `ensureAccessToken()` (tries `refreshAccess()` if no token), redirects to `/login` if still null. `component`: `<Outlet />`.
- `frontend/src/routes/_auth.dashboard.tsx` — `createFileRoute('/_auth/dashboard')`. Stub: `<h1>Dashboard</h1>` with user name from `useAuth`. Protected, all roles.
- `frontend/src/routes/_auth.staff.tsx` — `createFileRoute('/_auth/staff')`. Owner-only staff management page. Lists users; form to add cashier/kitchen; delete button.

### To modify

- `backend/cmd/api/main.go` — add DB open, wire `UserRepo → AuthUsecase, UserUsecase → AuthHandler, UserHandler → Chi router with RequireAuth middleware`
- `backend/go.mod` / `go.sum` — after `go get` of new deps
- `frontend/package.json` / `pnpm-lock.yaml` — after `pnpm add @hookform/resolvers`
- `frontend/src/main.tsx` — add login + `_auth` routes to `routeTree`; add `<Toaster />` for error toasts
- `memory-bank/api-contracts.md` — already has §Auth and §Users; mark as implemented
- `memory-bank/database-schema.md` — already contains the full schema; mark 001_init.sql committed
- `memory-bank/progress.md` — tick all M1 items

## Task breakdown

Dependencies flow top-to-bottom. Each numbered step is one atomic commit.

1. **Install Go deps** — from `backend/`: `go get golang-jwt/jwt/v5 golang.org/x/crypto gorm.io/gorm gorm.io/driver/postgres github.com/google/uuid github.com/go-playground/validator/v10`. Commit: `chore(deps): add auth + db deps to go.mod`.

2. **Install frontend deps** — from `frontend/`: `pnpm add @hookform/resolvers`. Commit: `chore(deps): add hookform resolvers`.

3. **`001_init.sql`** — full schema. Trigger function first, then all 8 tables, then all indexes. Verify locally with `make migrate` (requires docker compose up postgres). Commit: `feat(db): add 001_init.sql — full initial schema`.

4. **`002_seed_owner.sql`** — comment-only placeholder. Commit: `feat(db): add 002_seed_owner.sql placeholder`.

5. **`pkg/response`** — `response.go`. No deps outside stdlib. Commit: `feat(pkg): add response helpers`.

6. **`pkg/db`** — `db.go`. Commit: `feat(pkg): add db.Open helper`.

7. **`pkg/jwt` + tests** — `jwt.go` + `jwt_test.go`. Run `go test ./pkg/jwt/...`. Commit: `feat(pkg): add jwt Sign/Verify with tests`.

8. **Domain layer** — `domain/errors.go`, `domain/constants.go`, `domain/user.go`. Zero external imports. Commit: `feat(domain): add User entity, UserRepository interface, errors, constants`.

9. **`repository/user_repo.go`** — GORM implementation of `domain.UserRepository`. Commit: `feat(repo): add UserRepo GORM implementation`.

10. **`usecase/auth.go` + tests** — `AuthUsecase` with all 5 methods. Tests: register-owner-twice returns `ErrOwnerExists`, login with bad password returns `ErrInvalidCredentials`, refresh with replayed JTI returns `ErrInvalidCredentials`, `Me` returns user. Commit: `feat(usecase): add AuthUsecase with tests`.

11. **`usecase/user.go`** — `UserUsecase` list/create/update/soft-delete. Commit: `feat(usecase): add UserUsecase`.

12. **`middleware/jwt_auth.go`** — reads `Authorization: Bearer`, calls `pkg/jwt.Verify`, injects context values, checks role list. Commit: `feat(middleware): add RequireAuth JWT middleware`.

13. **`handler/auth_handler.go`** — 5 thin handler methods, cookie helpers. Commit: `feat(handler): add AuthHandler — 5 auth endpoints`.

14. **`handler/user_handler.go`** — 4 thin handler methods, owner-gated. Commit: `feat(handler): add UserHandler — owner CRUD`.

15. **Wire `main.go`** — open DB, build repo → usecase → handler, mount Chi routes, inject validator. Confirm `curl localhost:8080/healthz` still returns 200 and `curl -X POST localhost:8080/api/auth/register-owner` doesn't 404. Commit: `feat(api): wire auth + user routes in main.go`.

16. **`frontend/src/types/api.ts`** — TypeScript types mirroring Go DTOs. Commit: `feat(frontend): add API types`.

17. **`frontend/src/lib/currency.ts`** — `formatIDR`. Commit: `feat(frontend): add formatIDR currency helper`.

18. **`frontend/src/lib/auth.ts`** — memory token store + `refreshAccess`. Commit: `feat(frontend): add auth memory token store`.

19. **`frontend/src/lib/api.ts`** — `apiFetch` with auth injection + 401 retry. Commit: `feat(frontend): add apiFetch with auth injection`.

20. **`frontend/src/hooks/useAuth.ts`** — TanStack Query hooks. Commit: `feat(frontend): add useAuth hook`.

21. **`frontend/src/routes/login.tsx`** — login + register-owner form, Zod schemas, RHF, Radix Toast on error. Commit: `feat(frontend): add /login route`.

22. **`frontend/src/routes/_auth.tsx` + dashboard + staff** — protected layout, dashboard stub, staff management page. Commit: `feat(frontend): add protected layout, dashboard, staff routes`.

23. **Wire `main.tsx`** — add all new routes to route tree, add `<Toaster />`. Commit: `feat(frontend): wire auth routes into router`.

24. **Memory bank** — tick all M1 items in `progress.md`; confirm `api-contracts.md` and `database-schema.md` have no drift. Commit: `docs(memory-bank): mark M1 items complete`.

## Pseudocode — refresh rotation (trickiest function)

```
// internal/usecase/auth.go

func (a *AuthUsecase) Refresh(ctx context.Context, rawRefreshToken string) (accessToken, newRefreshToken string, err error) {
    claims, err := jwtpkg.Verify(rawRefreshToken, a.cfg.RefreshSecret)
    if err != nil {
        return "", "", domain.ErrInvalidCredentials  // expired or tampered
    }

    user, err := a.repo.FindByID(ctx, claims.Sub)
    if err != nil {
        return "", "", fmt.Errorf("usecase.Refresh: %w", err)
    }
    if user.DeletedAt != nil {
        return "", "", domain.ErrInvalidCredentials  // soft-deleted staff
    }

    // Replay / revocation guard — JTI must match what's stored
    if user.LastRefreshJTI == nil || *user.LastRefreshJTI != claims.JTI {
        return "", "", domain.ErrInvalidCredentials
    }

    newJTI := uuid.New().String()
    accessToken = jwtpkg.Sign(jwtpkg.Claims{Sub: user.ID, Role: user.Role, JTI: uuid.New().String()}, a.cfg.AccessSecret, 15*time.Minute)
    newRefreshToken = jwtpkg.Sign(jwtpkg.Claims{Sub: user.ID, JTI: newJTI}, a.cfg.RefreshSecret, 7*24*time.Hour)

    if err := a.repo.SetLastRefreshJTI(ctx, user.ID, newJTI); err != nil {
        return "", "", fmt.Errorf("usecase.Refresh set jti: %w", err)
    }
    return accessToken, newRefreshToken, nil
}
```

```
// frontend/src/lib/auth.ts

let _accessToken: string | null = null

export const getAccessToken = () => _accessToken
export const setAccessToken = (t: string) => { _accessToken = t }
export const clearAccessToken = () => { _accessToken = null }

export async function refreshAccess(): Promise<string | null> {
    try {
        const res = await fetch('/api/auth/refresh', { method: 'POST', credentials: 'include' })
        if (!res.ok) { clearAccessToken(); return null }
        const body = await res.json()
        setAccessToken(body.data.accessToken)
        return body.data.accessToken
    } catch {
        clearAccessToken()
        return null
    }
}
```

## Validation plan

```bash
# After each atomic commit:
make lint                # golangci-lint + eslint + prettier --check

# After step 7 (jwt pkg):
go test ./pkg/jwt/... -v

# After step 10 (auth usecase):
go test ./internal/usecase/... -v

# After step 15 (main.go wired):
make dev                          # docker compose up
curl localhost:8080/healthz       # → {"status":"ok"}
curl -s -X POST localhost:8080/api/auth/register-owner \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@brew.ly","password":"secret123","name":"Owner"}' | jq .
# → {success:true, data:{user}}

curl -s -X POST localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' -c /tmp/brew.jar \
  -d '{"email":"owner@brew.ly","password":"secret123"}' | jq .
# → {success:true, data:{accessToken, user}}; Set-Cookie: BrewlyRefresh= ...

ACCESS=$(curl -s -X POST localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' -c /tmp/brew.jar \
  -d '{"email":"owner@brew.ly","password":"secret123"}' | jq -r .data.accessToken)
curl -s localhost:8080/api/auth/me -H "Authorization: Bearer $ACCESS" | jq .
# → {success:true, data:{user}}

curl -s -X POST localhost:8080/api/auth/refresh -b /tmp/brew.jar -c /tmp/brew.jar | jq .
# → {success:true, data:{accessToken}} and new BrewlyRefresh cookie

# After step 23 (frontend wired):
# Open http://localhost:5173/login
# "Register as owner" button visible → fill form → redirects to /dashboard
# Log out → redirected to /login, old refresh cookie cleared
# Reload /dashboard → stays on /dashboard (silent refresh succeeds)
# Tamper cookie → redirected to /login
```

## Out of scope

- Password reset (owner can drop to `make psql` for v1; cashier accounts the owner re-creates).
- Email verification (no email infra in v1).
- 2FA, OAuth/social login.
- Rate limiting on `/api/auth/login` (deferred to v1.1 hardening pass).
- Full staff management UI with invite flow.
- `useUsers` hook and `_auth/staff` more than a stub (M2 picks it up once tables/menu exist).

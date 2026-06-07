# PRP — Auth (M1)

## Goal
Ship staff authentication and authorization end-to-end: an empty database lets the first visitor self-register as `owner`; existing staff log in to receive a short-lived access JWT plus an httpOnly refresh cookie; backend middleware gates protected endpoints by role. The owner can create cashier and kitchen accounts from a dashboard page. This is the foundation every subsequent milestone depends on.

## Context
- Architecture per `memory-bank/architecture.md`: clean layers, errors wrap on layer crossings, responses via `pkg/response`.
- Single-tenant, three roles: `owner`, `cashier`, `kitchen`. Defined in `internal/domain/constants.go`.
- Token strategy per `docs/ADR/ADR-004-table-token-security.md` for table tokens; staff JWT is separate but follows the same envelope discipline.
- Staff access token: HS256, 15 min, claims `{sub, role, iat, exp, jti}`, signed with `JWT_SECRET`.
- Refresh token: HS256, 7 days, signed with `REFRESH_SECRET`, delivered as cookie: `Path=/api/auth; HttpOnly; Secure (prod); SameSite=Lax`.
- Frontend never stores access token in localStorage. `src/lib/auth.ts` keeps it in a module-scoped variable; on app start the SPA calls `/api/auth/refresh` to get one.
- `users` table per `memory-bank/database-schema.md` already drafted in `001_init.sql`.

## File structure

### To create
- `backend/migrations/001_init.sql` — full schema (users + every table — this is M1's window to commit the entire initial schema)
- `backend/migrations/002_seed_owner.sql` — empty file with a comment; M1 doesn't actually seed since we self-register
- `backend/internal/domain/user.go` — `User` struct + `UserRepository` interface
- `backend/internal/domain/errors.go` — `ErrUserNotFound`, `ErrEmailTaken`, `ErrInvalidCredentials`, `ErrOwnerExists`
- `backend/internal/domain/constants.go` — `RoleOwner`, `RoleCashier`, `RoleKitchen`, refresh cookie name `BrewlyRefresh`
- `backend/internal/repository/user_repo.go` — GORM implementation
- `backend/internal/usecase/auth.go` — `AuthUsecase` with `RegisterOwner`, `Login`, `Refresh`, `Logout`, `Me`
- `backend/internal/usecase/user.go` — owner-scoped `UserUsecase` (list, create, patch, delete)
- `backend/internal/handler/auth_handler.go` — routes registered under `/api/auth`
- `backend/internal/handler/user_handler.go` — routes registered under `/api/users`
- `backend/internal/middleware/jwt_auth.go` — `RequireAuth(roles ...string) func(http.Handler) http.Handler`
- `backend/pkg/jwt/jwt.go` — `Sign`, `Verify` for both access and refresh; takes secret + ttl
- `backend/pkg/jwt/jwt_test.go`
- `backend/internal/usecase/auth_test.go`
- `frontend/src/routes/login.tsx` — login form with "Register as owner" button visible when `GET /api/auth/me` returns `404 OwnerNotExists`
- `frontend/src/routes/_auth.tsx` — root protected layout; `beforeLoad` calls `useAuth().ensureAccessToken()`
- `frontend/src/routes/_auth/staff/index.tsx` — owner-only staff management page
- `frontend/src/hooks/useAuth.ts`
- `frontend/src/hooks/useUsers.ts`
- `frontend/src/lib/auth.ts` — memory store + refresh call

### To modify
- `backend/cmd/api/main.go` — wire AuthUsecase, UserUsecase, jwt_auth middleware
- `backend/internal/handler/router.go` — mount `/api/auth` and `/api/users`
- `memory-bank/api-contracts.md` — these endpoints already listed under §Auth and §Users
- `memory-bank/database-schema.md` — already lists `users`
- `memory-bank/progress.md` — tick M1 items

## Task breakdown

1. **Schema** — write `001_init.sql` covering every M1–M4 table (we ship the full schema in M1 to avoid touching it later). Migrate locally with `make migrate`.
2. **JWT helper** — `pkg/jwt/jwt.go`: `Sign(claims, secret, ttl)`, `Verify(token, secret)`. Use `golang-jwt/jwt/v5`. Test coverage on tampered token, expired token, valid token.
3. **Domain layer** — `domain/user.go`, `domain/errors.go`, `domain/constants.go`.
4. **Repository** — `repository/user_repo.go` implementing `domain.UserRepository`. Use GORM. Methods: `Create`, `FindByEmail`, `FindByID`, `Exists`, `ExistsByRole`, `Update`, `SoftDelete`, `List`.
5. **Auth usecase** — `usecase/auth.go`. Methods:
   - `RegisterOwner` — fails with `ErrOwnerExists` if any owner exists (`repo.ExistsByRole(RoleOwner)`).
   - `Login` — looks up email, verifies bcrypt, issues access + refresh.
   - `Refresh` — verifies refresh token, rotates (issues new refresh, updates `users.last_refresh_jti`), returns new access token.
   - `Logout` — clears `last_refresh_jti`.
   - `Me` — returns user by id.
6. **User usecase** — `usecase/user.go` for owner CRUD on cashier/kitchen accounts.
7. **JWT middleware** — `middleware/jwt_auth.go`. Reads `Authorization: Bearer …`, validates, injects user id + role into context, gates by role list.
8. **Handlers** — `auth_handler.go` and `user_handler.go`. Thin. Set/clear refresh cookie at the boundary. Map domain errors to status codes.
9. **Wire main** — config load, db open, build repo → usecase → handler → router, start server.
10. **Frontend lib/auth.ts** — module-scope variable for access token, `setAccessToken`, `getAccessToken`, `clearAccessToken`. `refreshAccess()` calls `POST /api/auth/refresh` with `credentials: 'include'`.
11. **`useAuth` hook** — wraps `/me`, `/login`, `/logout`, `/register-owner`; manages access token via lib/auth.
12. **Routes** — `login.tsx` (toggles between login and register-owner based on `/me` response), `_auth.tsx` (beforeLoad redirects to /login), `_auth/staff/index.tsx`.
13. **Memory bank** — tick `progress.md` items.
14. **Manual smoke**: bring up `make dev`, hit `/login`, register owner, log in, hit dashboard, log out, log in again. Restart backend, confirm refresh cookie revives the session.

## Pseudocode (refresh rotation)

```
// usecase/auth.go
func (u *AuthUsecase) Refresh(ctx, oldRefreshToken) (newAccess string, newRefresh string, err error) {
    claims, err := jwt.Verify(oldRefreshToken, u.cfg.RefreshSecret)
    if err != nil { return "", "", ErrInvalidCredentials }

    user, err := u.repo.FindByID(ctx, claims.Sub)
    if err != nil { return "", "", err }

    if user.LastRefreshJTI != claims.JTI { // replay or revoked
        return "", "", ErrInvalidCredentials
    }

    newJTI := uuid.New().String()
    newAccess  = jwt.Sign(accessClaims(user), u.cfg.AccessSecret, 15*time.Minute)
    newRefresh = jwt.Sign(refreshClaims(user, newJTI), u.cfg.RefreshSecret, 7*24*time.Hour)

    if err := u.repo.SetLastRefreshJTI(ctx, user.ID, newJTI); err != nil {
        return "", "", fmt.Errorf("usecase.Refresh: %w", err)
    }
    return
}
```

## Validation plan
- `make lint` clean.
- `make test-backend` — auth_test covers: register-owner-twice fails, login wrong password, refresh rotation invalidates old, role gate rejects wrong role.
- `make test-frontend` — useAuth hook unit test (MSW-mocked) for login/refresh/logout.
- Manual smoke per task 14.

## Out of scope
- Password reset (deferred — single-owner deploys can drop into psql; cashier/kitchen accounts owner re-creates).
- Email verification (no email infra in v1).
- 2FA.
- OAuth/social login.

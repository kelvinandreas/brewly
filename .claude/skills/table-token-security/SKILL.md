---
name: table-token-security
description: Brewly table-token threat model and implementation. JWT structure with tvr claim, token_version invalidation, 4h TTL, per-token rate limiting, never store client-side. Triggers when working on middleware, table token generation, QR code generation, or customer endpoints.
---

## Threat model

The customer surface is anonymous. We protect against:
1. Remote / screenshot use of a leaked QR.
2. Spam song requests.
3. Forged tokens (signature check covers this).

We deliberately accept:
- TTL window (≤ 4h) where a leaked token still works. Owners regenerate weekly or after high-traffic events.

## Token structure

HS256, signed with `TABLE_TOKEN_SECRET`. Claims:

```json
{
  "tid": "<table uuid>",
  "tvr": "<int — copied from tables.token_version at sign time>",
  "jti": "<random uuid — used for rate-limit bucketing>",
  "iat": 1717459200,
  "exp": 1717473600
}
```

## Middleware order

```go
r.Route("/api/customer", func(r chi.Router) {
    r.Use(middleware.TableToken(repo, cfg))
    r.Use(middleware.SongRequestRateLimit(repo)) // reads token_jti from context
    r.Get("/menu", h.CustomerMenu)
    // ...
})
```

`TableToken` validation steps:
1. Read `Authorization: Bearer …` header.
2. Verify signature (HS256, `TABLE_TOKEN_SECRET`).
3. Check `exp`.
4. Look up `tables` by `tid`. Soft-deleted → 401.
5. Compare claim `tvr` to current `tables.token_version`. Mismatch → 401.
6. Inject `tableID`, `tokenJTI` into request context.

## Regeneration

Owner POST `/api/tables/:id/regenerate-token` →
```go
err := r.db.Model(&tableModel{}).
    Where("id = ?", tid).
    UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
```
Then re-sign a new token with the updated `tvr`, generate a new QR PNG, return both.

Old tokens fail step 5 next request.

## Rate limit

Song requests only: count rows with same `token_jti`:

```go
var count int64
if err := r.db.Model(&songRequestModel{}).
    Where("token_jti = ?", jti).
    Count(&count).Error; err != nil { return err }
if count >= 3 { return domain.ErrSongRateLimit }
```

## Client-side storage

```tsx
// routes/table/$tableId.tsx
const token = new URLSearchParams(window.location.search).get('token');
if (token) {
  setTableToken(token);                            // memory store in lib/auth.ts
  history.replaceState(null, '', `/table/${tableId}`); // strip from URL
}
```

Never write the token to localStorage or sessionStorage. Page refresh = back to QR scan. This is correct — the customer is in front of the QR.

## DO
- Always use `pkg/jwt` to sign — secret + algorithm + ttl handled centrally.
- Always compare `tvr` on every request.

## DON'T
- Don't cache table rows in middleware in v1 (a DB read per request is fine at cafe scale).
- Don't put `tid` in URL path AND token — must come from token only (else a guest could change `/table/3` to `/table/4`).
- Don't store the token in localStorage "for convenience" — defeats screenshot mitigation.

## Test cases

- Forged signature → 401
- Expired → 401
- `tvr` < current → 401 (regenerated mid-session)
- Soft-deleted table → 401
- 4th song request → 429
- Valid token, valid order → 200

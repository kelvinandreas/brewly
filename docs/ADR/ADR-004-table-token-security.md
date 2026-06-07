# ADR-004 — Table-token security: signed JWT + per-table token_version

## Context
The customer surface is anonymous. Anyone with the URL can hit our API. We must prevent:
1. Fake orders submitted from outside the cafe (someone screenshots a QR and orders remotely or repeatedly).
2. Spam song requests.
3. A leaked QR from being usable forever.

Constraints: single-cafe deploy, no Redis, no per-request DB hit if avoidable, regeneration must be one click for the owner.

## Decision
**JWT signed table tokens with a per-table version claim.**

Claims:
```json
{
  "tid": "<table uuid>",
  "tvr": 3,
  "jti": "<random uuid>",
  "iat": 1717459200,
  "exp": 1717473600
}
```

- `tid` — the table the token grants access to.
- `tvr` — copied from `tables.token_version` at sign time.
- `jti` — JWT ID, used as the unit for rate limiting (count of song_requests with this `token_jti`).
- TTL: 4 hours from `iat`.
- Algorithm: HS256, signed with `TABLE_TOKEN_SECRET` (separate from staff `JWT_SECRET`).

**Validation flow (`middleware/table_token.go`):**
1. Parse + verify signature. Fail → 401.
2. Check `exp`. Fail → 401.
3. Lookup table by `tid`. Soft-deleted → 401.
4. Compare claim's `tvr` to current `tables.token_version`. Mismatch → 401.
5. Inject `table_id` and `token_jti` into request context.

**Invalidation:** Owner hits `POST /api/tables/:id/regenerate-token` → backend bumps `tables.token_version` by 1 → re-signs a token claim with the new version → returns the new token and QR PNG. Every existing token for that table is now `tvr` < current and rejected on the next request. No blacklist needed; old tokens become naturally invalid.

**Rate limit:** `SELECT count(*) FROM song_requests WHERE token_jti = $1`. Hit at song-request submission only. >= 3 → 429.

**Token never persists client-side:** The QR URL contains the token as a query param. On `/table/$tableId.tsx` mount, the React layer reads the token from `window.location.search`, stores it in a memory-only `lib/auth.ts` module, then calls `history.replaceState(null, '', '/table/<id>')` to strip it from the URL. Refresh = back to the QR scan.

## Consequences
- Pros: stateless verification (one DB read for `token_version` per request — could be cached if it ever becomes hot); revocation is O(1) (one UPDATE); no extra infra; horizontal scaling friendly.
- Cons: `tvr` lookup per request adds 1 SELECT; not significant for cafe-scale traffic. JWT can't be revoked before the `tvr` bump propagates (microseconds), acceptable.
- Threat coverage:
  - Remote screenshot of QR: still works until either 4h TTL expires or owner regenerates. We document this in `docs/specification.md` — owners should regenerate weekly or after a high-traffic event.
  - Token reuse for spam orders: same `tid` + same TTL window. Rate limit applies to songs; orders are not rate-limited per token in v1 (cashier acknowledges every QR order). If abuse appears, add a per-token cap of N pending orders.

# Brewly — Specification

## Problem

Small cafes pay $50–$200/month for cloud POS subscriptions that lock them into hardware, payment processors, and a feature surface they don't use. Brewly is a free, self-hosted alternative tailored for one cafe at a time — no SaaS, no per-seat fees, no vendor lock-in. The differentiator over similar open-source POS projects is a polished customer-facing QR flow that combines **ordering and song requests** into one anonymous interface.

## User stories per role

### Owner
- As an owner, on first install I create my owner account by self-registering, so I don't need a hidden bootstrap step.
- As an owner, I manage menu categories and items so my menu reflects what we sell today.
- As an owner, I create/edit/delete tables and download or print their QR codes.
- As an owner, when a QR is leaked or a paper copy was discarded, I regenerate that table's token in one click — old QRs stop working immediately.
- As an owner, I create cashier and kitchen staff accounts with role-scoped access.
- As an owner, I see daily/weekly/monthly revenue, best sellers, and hourly volume so I can plan staffing and promotions.

### Cashier
- As a cashier, I take walk-up orders at the POS terminal — I pick a table, build an order, and confirm.
- As a cashier, when a customer pays, I record the payment method and amount received.
- As a cashier, I see the live song queue and can mark songs as played or skip them.

### Kitchen
- As kitchen staff, I see every active order on a KDS board updated in real time.
- As kitchen staff, I move orders through `confirmed → preparing → ready` with one tap; the front-of-house sees the update immediately.

### Customer
- As a customer, I scan the QR on my table, browse the menu, and place an order — no account, no app.
- As a customer, I search for a song on YouTube and request it; the cafe limits me to 3 requests per session.
- As a customer, I see my session's recent orders on the same page so I can confirm what I asked for.

## Feature specs

### Order state machine

```
pending  ──► confirmed ──► preparing ──► ready ──► completed
   │             │
   └─────────────┴───► cancelled
```

- `pending`: customer QR order, awaiting cashier confirmation
- `confirmed`: cashier-built orders start here; QR orders move here when cashier acknowledges
- `preparing`: kitchen has it
- `ready`: kitchen marks done, awaiting delivery to table
- `completed`: handed to customer, fully paid
- `cancelled`: terminal; allowed from `pending`, `confirmed`, `preparing` only

`completed` requires at least one `payments` row whose summed `amount_minor` equals `orders.total_minor`.

### Payment recording

- Cashier picks method: `cash` / `qris` / `card`. For cash, also enters `received_minor` (used to compute change at the UI layer — DB just stores the values).
- Multiple `payments` rows per order allowed (split payments).
- Once `sum(payments.amount_minor) >= orders.total_minor`, cashier can transition order to `completed`.

### Song queue

- Customer searches → backend proxies to `youtube.googleapis.com/youtube/v3/search` with `part=snippet&type=video&maxResults=10&q=…`.
- Selecting a result POSTs a `song_request` row — backend rejects if the token's `token_jti` already has 3 rows.
- Cashier UI shows queue ordered by `created_at`. Marking `playing` is exclusive — only one `playing` at a time per cafe (enforced by checking before update; race acceptable in single-tenant).

## Non-goals (v1 deliberately excludes)

- Payment-gateway integration (Midtrans, Stripe, etc.)
- Inventory management
- Multi-cafe / multi-tenant
- Customer accounts, loyalty programs
- Receipt printing (planned post-v1)
- Inventory deduction on order completion
- Offline-first / PWA
- Native mobile apps
- Background tax calculation
- Email/SMS notifications

## Security model

See [ADR-004](./ADR/ADR-004-table-token-security.md) for the table-token threat model and design.

Staff auth:
- Access JWT: 15 min, `HS256`, claims `{sub, role, iat, exp}`. Signed with `JWT_SECRET`.
- Refresh: 7 days, signed separate secret (`REFRESH_SECRET`), delivered as httpOnly + secure + SameSite=lax cookie at `/api/auth`.
- Refresh rotates on use (old token's `jti` is recorded in `users.last_refresh_jti` — replays rejected).
- Logout: cookie cleared server-side; `last_refresh_jti` set to a sentinel.

Table token (customer):
- TTL 4h, signed with `TABLE_TOKEN_SECRET`, claims `{tid: tableID, tvr: token_version, jti, iat, exp}`.
- Middleware: verify sig → look up `tables.token_version` → reject if mismatch or soft-deleted.
- Token never persists client-side beyond memory. URL-injection on scan, then `history.replaceState(null, '', '/table/<id>')` strips it from URL.

Rate limiting:
- Per-token song requests counted via `SELECT count(*) FROM song_requests WHERE token_jti = $1`; >= 3 returns `429`.
- No global rate limiting in v1 (single-cafe traffic levels don't warrant).

## Performance expectations

- p95 API latency under 100ms for staff endpoints, under 200ms for customer endpoints (YouTube search excluded — bounded by Google).
- KDS SSE clients: up to 5 concurrent (kitchen tablets) per cafe. Negligible memory per connection.
- Realistic peak: 200 orders/day, 30 tables. Postgres on the same host handles this with stock settings.

# API contracts

Response envelope (every endpoint):

```json
// success
{ "success": true, "data": "<object|array>", "message": "<optional>" }
// failure
{ "success": false, "error": "<machine-readable code>", "details": ["<optional field errors>"] }
```

Auth columns: `JWT` = staff access token bearer header; `Cookie` = httpOnly refresh cookie; `TableToken` = customer table-token bearer header; `—` = unauthenticated.

## Auth

| Method | Path | Auth | Body | Response data |
|---|---|---|---|---|
| POST | `/api/auth/register-owner` | — | `{email, password, name}` | `{user}` (only if no owner exists) |
| POST | `/api/auth/login` | — | `{email, password}` | `{accessToken, user}` + sets refresh cookie |
| POST | `/api/auth/refresh` | Cookie | — | `{accessToken}` + rotated refresh cookie |
| POST | `/api/auth/logout` | JWT | — | `{}` + clears refresh cookie |
| GET | `/api/auth/me` | JWT | — | `{user}` |

## Users (owner only)

| Method | Path | Body | Response data |
|---|---|---|---|
| GET | `/api/users` | — | `[user]` |
| POST | `/api/users` | `{email, password, name, role}` | `{user}` |
| PATCH | `/api/users/:id` | `{name?, role?, password?}` | `{user}` |
| DELETE | `/api/users/:id` | — | `{}` |

## Categories (owner)

| Method | Path | Body |
|---|---|---|
| GET | `/api/categories` | — |
| POST | `/api/categories` | `{name, displayOrder?}` |
| PATCH | `/api/categories/:id` | `{name?, displayOrder?}` |
| DELETE | `/api/categories/:id` | — |

## Menu items (owner)

| Method | Path | Body |
|---|---|---|
| GET | `/api/menu-items?categoryId=&availableOnly=` | — |
| POST | `/api/menu-items` | `{categoryId, name, description?, priceMinor, imageUrl?, isAvailable?}` |
| PATCH | `/api/menu-items/:id` | partial |
| DELETE | `/api/menu-items/:id` | — (soft) |

## Tables (owner)

| Method | Path | Body | Response data |
|---|---|---|---|
| GET | `/api/tables` | — | `[table]` |
| POST | `/api/tables` | `{label}` | `{table, qrToken, qrUrl}` |
| PATCH | `/api/tables/:id` | `{label?}` | `{table}` |
| DELETE | `/api/tables/:id` | — | `{}` |
| POST | `/api/tables/:id/regenerate-token` | — | `{qrToken, qrUrl}` |
| GET | `/api/tables/:id/qr.png` | — | `image/png` bytes |

## Orders (staff)

| Method | Path | Auth | Body |
|---|---|---|---|
| GET | `/api/orders?status=&from=&to=` | JWT | — |
| GET | `/api/orders/:id` | JWT | — |
| POST | `/api/orders` | JWT (cashier/owner) | `{tableId, items: [{menuItemId, quantity}], note?}` |
| PATCH | `/api/orders/:id/status` | JWT | `{status}` (state-machine validated) |
| POST | `/api/orders/:id/cancel` | JWT | `{reason?}` |

## Payments (staff)

| Method | Path | Body |
|---|---|---|
| POST | `/api/orders/:id/payments` | `{method, amountMinor, receivedMinor}` |
| GET | `/api/orders/:id/payments` | — |

## Reports (owner)

| Method | Path | Query |
|---|---|---|
| GET | `/api/reports/revenue` | `granularity=day|week|month`, `from`, `to` |
| GET | `/api/reports/best-sellers` | `from`, `to`, `limit?` |
| GET | `/api/reports/hourly-volume` | `date` |

## Song queue (staff)

| Method | Path | Body |
|---|---|---|
| GET | `/api/song-requests?status=` | — |
| PATCH | `/api/song-requests/:id/status` | `{status: playing|played|skipped}` |

## SSE streams

| Path | Auth | Events |
|---|---|---|
| `/api/sse/kitchen` | JWT (kitchen/cashier/owner) | `order.created`, `order.status_changed`, `order.cancelled` |
| `/api/sse/song-queue` | JWT (cashier/owner) | `song.requested`, `song.status_changed` |

Headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`, 15s keep-alive comment, client uses `EventSource` with `withCredentials: true` (refresh cookie attached for re-auth if needed).

## Customer endpoints (require `TableToken`)

| Method | Path | Body |
|---|---|---|
| GET | `/api/customer/menu` | — — returns categories + available items |
| POST | `/api/customer/orders` | `{items: [{menuItemId, quantity}], note?}` |
| GET | `/api/customer/orders/mine` | — — last 5 orders for this table in the token's session window |
| GET | `/api/customer/youtube/search?q=` | — — proxied YouTube Data API v3 search |
| POST | `/api/customer/song-requests` | `{youtubeVideoId, title, channelName, thumbnailUrl, note?}` |

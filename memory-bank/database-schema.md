# Database schema

Source of truth: numbered SQL files in `backend/migrations/`. Update this file after every applied migration. Format per table: columns, indexes, constraints, notes.

**Applied migrations:** `001_init.sql` (M1 — full schema, all 8 tables), `002_seed_owner.sql` (placeholder)

## `users`

| Column | Type | Constraints | Description |
|---|---|---|---|
| id | uuid | PK, default `gen_random_uuid()` | |
| email | text | UNIQUE, NOT NULL | Used as login |
| password_hash | text | NOT NULL | bcrypt cost 12 |
| name | text | NOT NULL | |
| role | text | NOT NULL, CHECK (`owner` / `cashier` / `kitchen`) | |
| last_refresh_jti | text | NULL | Refresh token rotation guard |
| created_at | timestamptz | NOT NULL, default `now()` | |
| updated_at | timestamptz | NOT NULL, default `now()` | Auto-updated by trigger |
| deleted_at | timestamptz | NULL | Soft delete |

Indexes: `idx_users_email_active` on `(email) WHERE deleted_at IS NULL`.

## `tables`

| Column | Type | Constraints | Description |
|---|---|---|---|
| id | uuid | PK | |
| label | text | NOT NULL | "1", "T1", "Patio-A" |
| token_version | int | NOT NULL, default 1 | Bumped to invalidate old QR tokens |
| created_at | timestamptz | NOT NULL, default `now()` | |
| updated_at | timestamptz | NOT NULL, default `now()` | |
| deleted_at | timestamptz | NULL | |

Indexes: `idx_tables_label_active` on `(label) WHERE deleted_at IS NULL`.

## `categories`

| Column | Type | Constraints |
|---|---|---|
| id | uuid | PK |
| name | text | NOT NULL |
| display_order | int | NOT NULL, default 0 |
| created_at | timestamptz | NOT NULL |
| updated_at | timestamptz | NOT NULL |
| deleted_at | timestamptz | NULL |

## `menu_items`

| Column | Type | Constraints |
|---|---|---|
| id | uuid | PK |
| category_id | uuid | FK → categories(id), NOT NULL |
| name | text | NOT NULL |
| description | text | NULL |
| price_minor | bigint | NOT NULL, CHECK (>= 0) — IDR minor units (cents), avoid float |
| image_url | text | NULL |
| is_available | bool | NOT NULL, default TRUE |
| created_at | timestamptz | NOT NULL |
| updated_at | timestamptz | NOT NULL |
| deleted_at | timestamptz | NULL |

Indexes: `idx_menu_items_category_active` on `(category_id) WHERE deleted_at IS NULL`.

## `orders`

| Column | Type | Constraints |
|---|---|---|
| id | uuid | PK |
| table_id | uuid | FK → tables(id), NOT NULL |
| status | text | NOT NULL, CHECK (`pending`, `confirmed`, `preparing`, `ready`, `completed`, `cancelled`) |
| source | text | NOT NULL, CHECK (`customer_qr`, `cashier_pos`) |
| total_minor | bigint | NOT NULL, CHECK (>= 0) |
| note | text | NULL |
| created_by_user_id | uuid | FK → users(id), NULL — set when `source = cashier_pos` |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL |
| deleted_at | timestamptz | NULL |

Indexes: `idx_orders_status_created` on `(status, created_at DESC) WHERE deleted_at IS NULL`, `idx_orders_table_created` on `(table_id, created_at DESC)`.

## `order_items`

| Column | Type | Constraints |
|---|---|---|
| id | uuid | PK |
| order_id | uuid | FK → orders(id), ON DELETE CASCADE, NOT NULL |
| menu_item_id | uuid | FK → menu_items(id), NOT NULL |
| name_snapshot | text | NOT NULL — captured at order time so renames don't rewrite history |
| price_minor_snapshot | bigint | NOT NULL |
| quantity | int | NOT NULL, CHECK (> 0) |
| created_at | timestamptz | NOT NULL |
| updated_at | timestamptz | NOT NULL |

## `payments`

| Column | Type | Constraints |
|---|---|---|
| id | uuid | PK |
| order_id | uuid | FK → orders(id), NOT NULL |
| method | text | NOT NULL, CHECK (`cash`, `qris`, `card`) |
| amount_minor | bigint | NOT NULL, CHECK (>= 0) |
| received_minor | bigint | NOT NULL, CHECK (>= 0) — for cash change calculation |
| recorded_by_user_id | uuid | FK → users(id), NOT NULL |
| created_at | timestamptz | NOT NULL |

Index: `idx_payments_order` on `(order_id)`.

## `song_requests`

| Column | Type | Constraints |
|---|---|---|
| id | uuid | PK |
| table_id | uuid | FK → tables(id), NOT NULL |
| token_jti | text | NOT NULL — JWT ID claim of the table token that made the request |
| youtube_video_id | text | NOT NULL |
| title | text | NOT NULL |
| channel_name | text | NOT NULL |
| thumbnail_url | text | NOT NULL |
| note | text | NULL |
| status | text | NOT NULL, CHECK (`queued`, `playing`, `played`, `skipped`) |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL |

Indexes: `idx_song_requests_token_jti` on `(token_jti)` — used by rate limiter; `idx_song_requests_status_created` on `(status, created_at)` — used by queue display.

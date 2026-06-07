-- 001_init.sql: full initial schema for Brewly
-- Forward-only. Never edit once committed.

-- Enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Shared trigger function: auto-update updated_at on every row change
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$;

-- ─── users ────────────────────────────────────────────────────────────────────

CREATE TABLE users (
  id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  email             text        NOT NULL UNIQUE,
  password_hash     text        NOT NULL,
  name              text        NOT NULL,
  role              text        NOT NULL CHECK (role IN ('owner', 'cashier', 'kitchen')),
  last_refresh_jti  text        NULL,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  deleted_at        timestamptz NULL
);

CREATE TRIGGER trg_users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE UNIQUE INDEX idx_users_email_active
  ON users (email)
  WHERE deleted_at IS NULL;

-- ─── tables ───────────────────────────────────────────────────────────────────

CREATE TABLE tables (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  label         text        NOT NULL,
  token_version int         NOT NULL DEFAULT 1,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  deleted_at    timestamptz NULL
);

CREATE TRIGGER trg_tables_updated_at
  BEFORE UPDATE ON tables
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE UNIQUE INDEX idx_tables_label_active
  ON tables (label)
  WHERE deleted_at IS NULL;

-- ─── categories ───────────────────────────────────────────────────────────────

CREATE TABLE categories (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name          text        NOT NULL,
  display_order int         NOT NULL DEFAULT 0,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  deleted_at    timestamptz NULL
);

CREATE TRIGGER trg_categories_updated_at
  BEFORE UPDATE ON categories
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─── menu_items ───────────────────────────────────────────────────────────────

CREATE TABLE menu_items (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id   uuid        NOT NULL REFERENCES categories(id),
  name          text        NOT NULL,
  description   text        NULL,
  price_minor   bigint      NOT NULL CHECK (price_minor >= 0),
  image_url     text        NULL,
  is_available  bool        NOT NULL DEFAULT TRUE,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  deleted_at    timestamptz NULL
);

CREATE TRIGGER trg_menu_items_updated_at
  BEFORE UPDATE ON menu_items
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_menu_items_category_active
  ON menu_items (category_id)
  WHERE deleted_at IS NULL;

-- ─── orders ───────────────────────────────────────────────────────────────────

CREATE TABLE orders (
  id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  table_id            uuid        NOT NULL REFERENCES tables(id),
  status              text        NOT NULL CHECK (status IN ('pending', 'confirmed', 'preparing', 'ready', 'completed', 'cancelled')),
  source              text        NOT NULL CHECK (source IN ('customer_qr', 'cashier_pos')),
  total_minor         bigint      NOT NULL CHECK (total_minor >= 0),
  note                text        NULL,
  created_by_user_id  uuid        NULL REFERENCES users(id),
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  deleted_at          timestamptz NULL
);

CREATE TRIGGER trg_orders_updated_at
  BEFORE UPDATE ON orders
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_orders_status_created
  ON orders (status, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_orders_table_created
  ON orders (table_id, created_at DESC);

-- ─── order_items ──────────────────────────────────────────────────────────────

CREATE TABLE order_items (
  id                    uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id              uuid        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  menu_item_id          uuid        NOT NULL REFERENCES menu_items(id),
  name_snapshot         text        NOT NULL,
  price_minor_snapshot  bigint      NOT NULL,
  quantity              int         NOT NULL CHECK (quantity > 0),
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_order_items_updated_at
  BEFORE UPDATE ON order_items
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_order_items_order
  ON order_items (order_id);

-- ─── payments ─────────────────────────────────────────────────────────────────

CREATE TABLE payments (
  id                    uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id              uuid        NOT NULL REFERENCES orders(id),
  method                text        NOT NULL CHECK (method IN ('cash', 'qris', 'card')),
  amount_minor          bigint      NOT NULL CHECK (amount_minor >= 0),
  received_minor        bigint      NOT NULL CHECK (received_minor >= 0),
  recorded_by_user_id   uuid        NOT NULL REFERENCES users(id),
  created_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order
  ON payments (order_id);

-- ─── song_requests ────────────────────────────────────────────────────────────

CREATE TABLE song_requests (
  id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  table_id          uuid        NOT NULL REFERENCES tables(id),
  token_jti         text        NOT NULL,
  youtube_video_id  text        NOT NULL,
  title             text        NOT NULL,
  channel_name      text        NOT NULL,
  thumbnail_url     text        NOT NULL,
  note              text        NULL,
  status            text        NOT NULL CHECK (status IN ('queued', 'playing', 'played', 'skipped')),
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_song_requests_updated_at
  BEFORE UPDATE ON song_requests
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_song_requests_token_jti
  ON song_requests (token_jti);

CREATE INDEX idx_song_requests_status_created
  ON song_requests (status, created_at);

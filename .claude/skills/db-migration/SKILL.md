---
name: db-migration
description: Brewly migration conventions — file naming, forward-only, never DROP COLUMN, soft-delete pattern, updated_at trigger, index naming. Triggers when creating files in backend/migrations/ or writing GORM model definitions.
---

## File naming
- `NNN_<verb>_<thing>.sql` — three-digit zero-padded, then snake_case description.
- `001_init.sql`, `002_add_indexes_orders.sql`, `003_add_loyalty_field.sql`.
- One migration = one logical change. Don't bundle unrelated edits.

## Forward-only
- No down migrations in v1.
- Once merged, a migration file is immutable. Need a fix? Add the next-numbered file.

## Soft delete
- Every domain table has `deleted_at TIMESTAMPTZ`.
- Indexes that scan live rows: `WHERE deleted_at IS NULL`.

## updated_at trigger

`001_init.sql` defines once:

```sql
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

Every domain table attaches:

```sql
CREATE TRIGGER trg_<table>_updated_at
BEFORE UPDATE ON <table>
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

## Index naming
- `idx_<table>_<columns>[_<predicate>]`
- Examples: `idx_orders_status_created`, `idx_users_email_active`.

## DO

```sql
-- 005_add_order_note_index.sql
CREATE INDEX idx_orders_table_pending
  ON orders (table_id)
  WHERE status = 'pending' AND deleted_at IS NULL;
```

```sql
-- 006_add_user_phone.sql -- additive, nullable
ALTER TABLE users ADD COLUMN phone TEXT;
```

## DON'T

```sql
-- WRONG: drops data on existing rows
ALTER TABLE menu_items DROP COLUMN description;
```

```sql
-- WRONG: edits an already-merged migration
-- (touching 001_init.sql after merge)
```

```sql
-- WRONG: hard delete
DELETE FROM users WHERE email = 'x@y.com';
-- Use UPDATE users SET deleted_at = now() WHERE …
```

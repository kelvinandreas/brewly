# DB migration rules

Globs: `backend/migrations/*.sql`

## Naming
- Numeric prefix, zero-padded to 3: `001_init.sql`, `002_seed_owner.sql`, `003_…`.
- Verb-first description: `add_loyalty_field`, `add_indexes_orders`.

## Reversibility
- Forward-only. No down migrations in v1 (single cafe, low risk).
- Don't edit a merged migration — add a new one.

## Backward-compat
- Never `DROP COLUMN`. Add a new nullable column; deprecate later if needed.
- New `NOT NULL` columns need a default.

## Soft delete
- Every domain table has `deleted_at timestamptz NULL`.
- Indexes filter on `WHERE deleted_at IS NULL`.

## Triggers
- `updated_at` auto-updated via shared trigger function `set_updated_at()` declared in `001_init.sql`.

## Indexes
- Naming: `idx_<table>_<columns>[_<predicate>]`.
- Always index foreign keys and `(deleted_at)` predicate is implicit via partial index.

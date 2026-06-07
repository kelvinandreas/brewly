---
name: gorm-patterns
description: GORM conventions for Brewly — soft delete, preloading, N+1 avoidance, when to drop to raw SQL, hooks, UUID primary keys. Triggers when editing files in repository/ or GORM model structs.
---

## Models

- Tag PKs explicitly: `ID uuid.UUID \`gorm:"type:uuid;primaryKey;default:gen_random_uuid()"\``.
- Soft delete: embed `DeletedAt gorm.DeletedAt \`gorm:"index"\`` — GORM automatically scopes queries with `deleted_at IS NULL` and switches `Delete` to a soft delete.
- Repository model lives in `repository/` package, NOT `domain/`. Convert to/from domain entity at the repo boundary.

## Preload vs Joins

- `Preload` — issues a second query, easier to read, no row multiplication. Default choice for has-many.
- `Joins` — single query but rows multiply by has-many cardinality; you must `DISTINCT` or post-process. Use only when filtering on the joined table.

## N+1 prevention

If you loop and call `Find` per iteration, you have N+1. Either `Preload` upfront or batch with `IN`:

```go
var ids []uuid.UUID
for _, o := range orders { ids = append(ids, o.ID) }
var items []Item
db.Where("order_id IN ?", ids).Find(&items)
// then group in Go
```

## Raw SQL — when

Drop to `db.Raw(...).Scan(&result)` when:
- Reporting queries (window functions, GROUP BY across joins).
- Three or more chained ORM methods.
- You need a specific index hint.

Always pass parameters as args — never concatenate.

## Hooks
- `BeforeCreate` — only if logic must run for every direct insert. Prefer keeping logic in usecase.
- Never put domain logic in hooks; they're invisible to readers of the usecase.

## DO

```go
// soft delete configured
type orderModel struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Status    string
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// preload to avoid N+1
var orders []orderModel
err := r.db.WithContext(ctx).
    Preload("Items").
    Where("status = ?", "ready").
    Find(&orders).Error
```

```go
// raw SQL for reports
var rows []revenueRow
err := r.db.WithContext(ctx).Raw(`
    SELECT DATE_TRUNC('day', created_at) AS bucket,
           SUM(total_minor)              AS total
    FROM orders
    WHERE status = 'completed'
      AND created_at BETWEEN ? AND ?
      AND deleted_at IS NULL
    GROUP BY 1
    ORDER BY 1
`, from, to).Scan(&rows).Error
```

## DON'T

```go
// N+1
for _, o := range orders {
    var items []itemModel
    r.db.Where("order_id = ?", o.ID).Find(&items) // ← runs once per order
    o.Items = items
}
```

```go
// hard delete bypasses soft delete
r.db.Unscoped().Delete(&menuItem) // ← never; we preserve order history via name_snapshot
```

## UUID primary keys
- Postgres extension `pgcrypto` is enabled in `001_init.sql`.
- Default `gen_random_uuid()` runs server-side — no need to set IDs in Go.
- For test fixtures, generate explicitly: `uuid.New()`.

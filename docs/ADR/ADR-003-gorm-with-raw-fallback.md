# ADR-003 — GORM as default, raw SQL escape hatch for complex queries

## Context
The team is more productive with an ORM for CRUD-heavy work (menu, users, tables). But reporting queries — revenue by hour, best sellers across date ranges — read poorly through GORM's chained API and tend to generate non-obvious SQL.

## Decision
GORM is the default for all persistence (CRUD, simple joins, preloading). When a query is reporting-flavored or performance-sensitive, the repository function uses `db.Raw(...).Scan(&result)` with a hand-written SQL string. The repository interface in `domain/` is identical either way — the caller doesn't know which path was taken.

Concrete rule: **reach for raw SQL when the GORM expression includes a third chained method or a window function.**

## Consequences
- Pros: fast CRUD with safety (GORM soft delete, hooks); explicit SQL where it matters (reports).
- Cons: two patterns in one file; future contributors must know which to pick. The `gorm-patterns` skill documents the rule.
- Migrations stay hand-written SQL — GORM `AutoMigrate` is **never** used. Schema is owned by `backend/migrations/`.

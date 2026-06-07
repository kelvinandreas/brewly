# Brewly — AI Coding Rules

Brewly is a self-hosted, open-source POS for a single small cafe. Customers order and request songs anonymously through a QR-per-table; staff (owner / cashier / kitchen) sign in to a dashboard. Stack: Go + Chi + GORM + PostgreSQL backend; React 19 + TanStack Router + TanStack Query frontend; everything containerized.

## Stack

| Layer            | Tool                                                                           |
| ---------------- | ------------------------------------------------------------------------------ |
| Backend          | Go 1.25, Chi v5, GORM                                                          |
| Frontend         | React 19, Vite, TanStack Router/Query, Tailwind                                |
| Database         | PostgreSQL 16                                                                  |
| Auth             | JWT access (15m) + refresh httpOnly cookie (7d); customer table-token JWT (4h) |
| Real-time        | SSE (Server-Sent Events) — never WebSocket in v1                               |
| Containerization | Docker Compose                                                                 |

## Architecture — non-negotiable layer rules

- `handler/` parses input, calls a usecase, returns a response. **Zero** business logic.
- `usecase/` contains business logic. **Zero** HTTP or DB concerns. Depends only on interfaces from `domain/`.
- `domain/` declares entities and repository interfaces. **Zero** external imports — no GORM, no Chi, no JWT.
- `repository/` is the only place GORM appears. Repositories implement interfaces declared in `domain/`.
- Every error crossing a layer boundary is wrapped: `fmt.Errorf("usecase.CreateOrder: %w", err)`.
- Frontend mirrors this: `lib/api.ts` is the only place fetch appears; `hooks/` wraps TanStack Query; routes/components consume hooks.

## Critical rules — read before coding

Always load these into context before starting a task:

1. `memory-bank/architecture.md` — system diagram and data flow
2. `memory-bank/database-schema.md` — current schema; update after every migration
3. `memory-bank/api-contracts.md` — current API; update after every endpoint change
4. `memory-bank/progress.md` — what's done, what's in progress, what's next

When working on a feature, also load the matching `docs/ADR/*.md` if one exists.

## Development commands

```bash
make dev           # docker compose up — postgres + backend (Air) + frontend (Vite)
make migrate       # apply pending SQL migrations
make test          # backend `go test ./...` + frontend `pnpm test`
make lint          # golangci-lint + eslint + prettier --check
make commit        # interactive conventional-commit helper (scripts/commit.sh)
make build-prod    # build production images via docker-compose.prod.yml
```

## What NOT to do

- Don't put business logic in handlers.
- Don't import GORM outside `repository/`.
- Don't hard-delete rows — every domain table uses soft delete via `deleted_at`.
- Don't store tokens in `localStorage` — staff JWT lives in memory + httpOnly refresh cookie; table token lives only in React state for the browser session.
- Don't add `cafe_id` columns — this is single-tenant.
- Don't introduce Redis, message queues, or WebSockets in v1. SSE handles all real-time.

## Skills

Project skills in `.claude/skills/` auto-load by context — don't reference them manually.
Global skills (`session-audit`, `prp-generator`) install separately. See `docs/ENGINEERING.md` → "Setting up global Claude Code skills".

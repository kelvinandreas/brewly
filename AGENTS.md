# Brewly — Universal Agent Instructions

Brewly is an open-source POS for a single cafe. Staff dashboard + anonymous customer QR flow (order + song request). Backend: Go + Chi + GORM + PostgreSQL. Frontend: React 19 + TanStack Router/Query + Tailwind. Docker Compose for everything.

## Repo structure (top level)

- `backend/` — Go service. Clean architecture: `domain/` → `usecase/` → `repository/` + `handler/`. Migrations in `backend/migrations/`.
- `frontend/` — React 19 + Vite SPA. File-based routes in `src/routes/`. TanStack Query hooks in `src/hooks/`.
- `memory-bank/` — living context docs. Update after milestones.
- `docs/` — engineering guide, full spec, ADRs.
- `PRPs/` — Project Requirement Prompts. Use `PRPs/templates/prp_base.md` to plan new features.
- `.claude/` — Claude Code rules, commands, skills (project-scoped).
- `examples/` — canonical patterns for new files.

## Conventions

**Go backend**
- Module path `github.com/your-handle/brewly`. Go 1.23.
- Handlers: parse request → call usecase → write response via `pkg/response`. No business logic.
- Usecases depend on `domain.XxxRepository` interfaces, never on GORM.
- Errors: wrap with `fmt.Errorf("layer.Method: %w", err)`. Sentinel errors in `internal/domain/errors.go`.
- HTTP responses always go through `pkg/response.OK` / `pkg/response.Error` — never `http.Error` directly.
- Table-driven tests with `t.Run(name, …)`.

**React frontend**
- One component per file. Named + default export.
- Server data through TanStack Query hooks (`src/hooks/use*.ts`); never `fetch` in components.
- Routes are file-based. Protected routes nest under `_auth.tsx`.
- All currency display via `lib/currency.ts` (IDR, `id-ID`).
- Forms: React Hook Form + Zod schema in the same file.

**Database**
- UUID primary keys via `gen_random_uuid()`.
- Every table has `created_at`, `updated_at`, and (where applicable) `deleted_at` for soft delete.
- New migrations: next-numbered file in `backend/migrations/` (`004_…sql`). Never edit a merged migration.
- Never `DROP COLUMN` — add a new nullable column instead.

**Commits**
- Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`, `style:`.
- Imperative, present tense. No mention of AI tools, no emoji unless asked.
- Prefer `make commit` to drive the format check.

## What NOT to do

- Don't put business logic in handlers or business validation in components.
- Don't import GORM types into `domain/` or `usecase/`.
- Don't store any token in `localStorage` or `sessionStorage`.
- Don't add multi-tenant columns (`cafe_id` etc.) — single deployment, single cafe.
- Don't add Redis, Kafka, or WebSocket in v1. SSE only for real-time.
- Don't introduce a payment-gateway SDK — payments are recorded manually by the cashier.
- Don't bypass `lib/api.ts` for HTTP calls or `pkg/response` for responses.
- Don't commit `.env`. Update `.env.example` instead.

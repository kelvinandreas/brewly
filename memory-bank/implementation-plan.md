# Implementation plan

Each milestone is a coherent deliverable. After each one: tick `progress.md`, update `api-contracts.md` if endpoints changed, update `database-schema.md` if a migration was added.

| # | Milestone | Deliverables | Suggested PRPs |
|---|---|---|---|
| **M0** | Repo skeleton + vibe-engineering setup | All MD files: CLAUDE.md, AGENTS.md, README.md, memory-bank/*, docs/*, ADRs, .claude/*, PRPs/*, examples/*, configs, Docker, Makefile. Empty backend/ and frontend/ scaffolds that build and start cleanly. | `PRPs/m0-bootstrap.md` |
| **M1** | Auth + user management | `users` table, owner self-registration on empty DB, JWT access (15m) + refresh (7d httpOnly cookie), `/auth/*` endpoints, owner-only CRUD for cashier/kitchen users, `jwt_auth` middleware, frontend login route, memory token storage. | `PRPs/m1-auth.md` (use `PRPs/examples/prp_auth.md` as the template) |
| **M2** | Menu + table management + QR | Categories + menu items CRUD with soft delete, tables CRUD, server-side QR PNG generation, signed table-token JWT with `token_version`, `table_token` middleware, regenerate-QR endpoint, owner-dashboard pages for menu/tables. | `PRPs/m2-menu-tables.md` |
| **M3** | Ordering + payments + KDS (SSE) | Order/order_items/payment tables, cashier POS terminal route, customer ordering route (`/table/$tableId`), order status state machine, payment recording, kitchen route with SSE-driven KDS board, SSE endpoint with proper goroutine lifecycle. | `PRPs/m3-orders-kds.md` |
| **M4** | Song requests + reports | YouTube Data API v3 search, song request submission with per-token rate limit (3/lifetime, enforced in DB), song queue management for cashier/owner, reports dashboard (daily/weekly/monthly revenue, best sellers, hourly volume), date-range filter. | `PRPs/m4-songs-reports.md` |

A v1.0.0 tag is cut after M4. Anything beyond M4 (printer integration, inventory, customer loyalty) is explicitly out of scope and documented as such in `docs/specification.md` → Non-Goals.

## Milestone deliverable expansions

### M0 — Bootstrap

- Root MD files: CLAUDE.md, AGENTS.md, README.md, CONTRIBUTING.md, LICENSE
- All `memory-bank/` files
- All `docs/` files including 6 ADRs
- All `.claude/` (commands, rules, skills)
- All `PRPs/` (template + auth example + ai_docs README)
- All `examples/` files
- backend/: go.mod, cmd/api/main.go printing "ok" on `GET /healthz`, .air.toml, Dockerfile, Dockerfile.dev
- frontend/: package.json, vite.config.ts, tsconfig.json, eslint.config.js, prettier.config.js, tailwind.config.ts, postcss.config.cjs, Dockerfile, Dockerfile.dev, nginx.conf, src/main.tsx, src/routes/__root.tsx, src/routes/index.tsx
- Root: docker-compose.yml, docker-compose.prod.yml, .env.example, .gitignore, .dockerignore, .editorconfig, .golangci.yml, Makefile, scripts/commit.sh

### M1 — Auth

- migrations/001_init.sql (full v1 schema — every table)
- backend: domain.User + UserRepository, repository.UserRepo, usecase.AuthUsecase + UserUsecase, handler.AuthHandler + UserHandler, middleware.JWTAuth, pkg/jwt
- frontend: /login route, useAuth hook, lib/auth.ts (memory token store), _auth.tsx protected layout, /_auth/staff page for owner
- update memory-bank/api-contracts.md, database-schema.md, progress.md

### M2 — Menu, tables, QR

- backend: domain.Category, domain.MenuItem, domain.Table + repositories + usecases + handlers, middleware.TableToken, pkg/jwt (table token claims), QR PNG generator
- frontend: /_auth/menu/*, /_auth/tables/*, table token URL extraction in routes/table/$tableId.tsx

### M3 — Orders, payments, KDS

- backend: domain.Order, OrderItem, Payment + repos + usecases + handlers, order state machine, SSE broker, handler/sse_handler.go
- frontend: cashier POS terminal, customer order page, kitchen route with EventSource, payment form

### M4 — Songs, reports

- backend: pkg/youtube, song_request repo + usecase + handler, rate limit middleware, report usecase + handler (raw SQL for reports)
- frontend: customer song search/request UI, staff song queue, reports dashboard with date range filter

# Progress

Legend: `[ ]` todo · `[~]` in progress · `[x]` done

## M0 — Bootstrap

- [x] Root MD files (CLAUDE.md, AGENTS.md, README.md, CONTRIBUTING.md, LICENSE)
- [x] memory-bank/ files
- [x] docs/ENGINEERING.md + docs/specification.md
- [x] All 6 ADRs
- [x] .claude/commands, .claude/rules, .claude/skills
- [x] PRPs/ templates + auth example
- [x] examples/ canonical files
- [x] backend/ go.mod, main.go boots and responds 200 to `/healthz`
- [x] frontend/ pnpm install, `vite dev` serves the React shell
- [x] docker-compose.yml `make dev` brings up pg + backend + frontend cleanly
- [x] Makefile targets all execute

## M1 — Auth

- [x] migrations/001_init.sql committed (full schema — all 8 tables)
- [x] migrations/002_seed_owner.sql (placeholder; registration is self-serve via API)
- [x] backend: domain.User, repository.UserRepo, usecase.AuthUsecase, handler.AuthHandler
- [x] middleware.JWTAuth (RequireAuth with role enforcement)
- [x] frontend: /login route, useAuth hook, memory token store
- [x] update memory-bank/api-contracts.md
- [x] update memory-bank/database-schema.md

## M2 — Menu, tables, QR (pending)

## M3 — Orders, payments, KDS (pending)

## M4 — Songs, reports (pending)

# Progress

Legend: `[ ]` todo · `[~]` in progress · `[x]` done

## M0 — Bootstrap

- [x] Root MD files (CLAUDE.md, AGENTS.md, README.md, CONTRIBUTING.md, LICENSE)
- [x] memory-bank/ files
- [x] docs/ENGINEERING.md + docs/specification.md
- [x] All 6 ADRs
- [x] .claude/commands, .claude/rules, .claude/skills
- [x] PRPs/ templates + auth example
- [ ] examples/ canonical files
- [ ] backend/ go.mod, main.go boots and responds 200 to `/healthz`
- [ ] frontend/ pnpm install, `vite dev` serves the React shell
- [ ] docker-compose.yml `make dev` brings up pg + backend + frontend cleanly
- [ ] Makefile targets all execute

## M1 — Auth

- [ ] migrations/001_init.sql committed
- [ ] migrations/002_seed_owner.sql (idempotent — only inserts if no owner)
- [ ] backend: domain.User, repository.UserRepo, usecase.AuthUsecase, handler.AuthHandler
- [ ] middleware.JWTAuth
- [ ] frontend: /login route, useAuth hook, memory token store
- [ ] update memory-bank/api-contracts.md
- [ ] update memory-bank/database-schema.md

## M2 — Menu, tables, QR (pending)

## M3 — Orders, payments, KDS (pending)

## M4 — Songs, reports (pending)

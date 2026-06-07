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

## M2 — Menu, tables, QR

- [x] backend: domain.Category, domain.MenuItem, domain.Table (entities + repo interfaces)
- [x] backend: domain/errors.go additions (ErrCategoryNotFound, ErrMenuItemNotFound, ErrTableNotFound, ErrTableLabelTaken, ErrMenuItemUnavailable)
- [x] backend: domain/constants.go additions (ContextKeyTableID, ContextKeyTokenJTI)
- [x] backend: pkg/tabletoken — Sign/Verify, 4h TTL, claims {tid, tvr, jti}
- [x] backend: pkg/qrcode — Generate() returns 256×256 PNG bytes
- [x] backend: repository.CategoryRepo, repository.MenuItemRepo, repository.TableRepo (GORM)
- [x] backend: middleware.RequireTableToken (sig → exp → DB lookup → token_version check → ctx inject)
- [x] backend: usecase.CategoryUsecase, usecase.MenuItemUsecase, usecase.TableUsecase
- [x] backend: handler.CategoryHandler, handler.MenuItemHandler, handler.TableHandler, handler.CustomerHandler
- [x] backend: main.go wired (all M2 routes under correct auth middleware)
- [x] frontend: types/api.ts additions (Category, MenuItem, Table, QR types, CustomerMenu types)
- [x] frontend: lib/tableAuth.ts (module-scope table token store)
- [x] frontend: hooks/useCategories, hooks/useMenuItems, hooks/useTables
- [x] frontend: routes/_auth.menu.tsx (owner menu management, category + item CRUD)
- [x] frontend: routes/_auth.tables.tsx (table list, QR modal, regen token)
- [x] frontend: routes/table.$tableId.tsx (customer QR landing + menu browse)
- [x] frontend: main.tsx route wiring; dashboard nav links updated

## M3 — Orders, payments, KDS (pending)

## M4 — Songs, reports (pending)

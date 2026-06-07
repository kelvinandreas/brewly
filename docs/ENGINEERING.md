# Engineering guide

## Prerequisites

- Docker 25+ and Docker Compose v2
- Go 1.23+ (only if you want to run backend tests locally without Docker)
- Node 20 + pnpm 9 (only if you want to run the frontend dev server locally without Docker)
- `make`

## Local setup

```bash
git clone https://github.com/your-handle/brewly
cd brewly
cp .env.example .env
# Edit .env — at minimum: DB_PASSWORD, JWT_SECRET, REFRESH_SECRET, TABLE_TOKEN_SECRET, YOUTUBE_API_KEY
make dev
```

`make dev` runs `docker compose up`. First boot pulls images, runs migrations (entrypoint waits for pg), and starts:

- `postgres` on `:5432`
- `backend` on `:8080` (hot reload via Air)
- `frontend` on `:5173` (Vite)

Open `http://localhost:5173/login` → "Register owner" appears because the DB has no owner yet.

## Makefile targets

| Target | What it does |
|---|---|
| `make dev` | `docker compose up` — dev stack |
| `make dev-down` | `docker compose down` |
| `make migrate` | Apply pending SQL migrations against the running DB |
| `make migrate-new name=add_foo` | Scaffold `backend/migrations/NNN_add_foo.sql` |
| `make test` | `go test ./...` + `pnpm --filter frontend test` |
| `make test-backend` | backend tests only |
| `make test-frontend` | frontend tests only |
| `make lint` | `golangci-lint run` + `pnpm lint` + `pnpm format:check` |
| `make fmt` | `gofmt -w` + `pnpm format` |
| `make commit` | runs `scripts/commit.sh` — interactive conventional commit |
| `make build-prod` | builds prod images via `docker-compose.prod.yml` |
| `make logs` | tails docker compose logs |
| `make psql` | drops into psql against the dev DB |
| `make seed` | runs `scripts/seed.sh` (categories + sample items + tables for demos) |

## Migration workflow

1. `make migrate-new name=add_loyalty_field` → creates `backend/migrations/00X_add_loyalty_field.sql`.
2. Write SQL (additive only — see `db-migration` skill).
3. `make migrate` to apply locally.
4. **Update `memory-bank/database-schema.md`** in the same commit.
5. Commit: `feat(db): add loyalty_field to users`.

## Testing strategy

- **Backend unit**: usecase tests with mocked repositories (interfaces in `domain/`). Aim for usecase coverage over handler coverage.
- **Backend integration**: a `repository/*_test.go` per repo, runs against a `postgres:16-alpine` testcontainer. `t.Cleanup` truncates tables between cases.
- **Frontend**: Vitest + Testing Library for hooks (mock fetch with MSW) and presentational components. No e2e in v1.
- **Manual smoke**: `make seed` then walk the 3 flows (cashier place order, customer scan + order, kitchen mark ready).

## Linting

- Go: `.golangci.yml` enables errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, goimports, revive. Run with `golangci-lint run ./...`.
- TS: ESLint flat config in `frontend/eslint.config.js`. Prettier with `printWidth: 100`, single quotes, trailing commas.

## Setting up global Claude Code skills

Two personal-workflow skills live in your home directory, not in this repo:

```bash
mkdir -p ~/.claude/skills/session-audit
mkdir -p ~/.claude/skills/prp-generator
```

**`~/.claude/skills/session-audit/SKILL.md`**

```markdown
---
name: session-audit
description: End-of-session checklist. Activates when the user says "audit session", "end of session", or "what did we miss".
---

When the user invokes this, audit the current session:

1. **Schema drift** — did any migration get added under `backend/migrations/`? If so, is `memory-bank/database-schema.md` updated to match? If not, prompt to update.
2. **API drift** — did any handler get added/changed under `backend/internal/handler/`? Is `memory-bank/api-contracts.md` updated? Diff the endpoints and prompt for missing entries.
3. **Progress drift** — is `memory-bank/progress.md` reflective of what shipped this session?
4. **Decision drift** — did the session introduce a load-bearing architecture choice (new dependency, new pattern, abandoning a documented approach)? If yes, propose an ADR filename and outline.
5. **Skill gaps** — did you find yourself re-explaining a rule to the user? That's a candidate for a new project skill in `.claude/skills/`.

Output: a checklist with `[ ]` items the user should fix before stopping work. Be specific — name files, line ranges, and exact additions needed.
```

**`~/.claude/skills/prp-generator/SKILL.md`**

```markdown
---
name: prp-generator
description: Expand a rough feature description into a PRP document. Activates when the user says "generate PRP for X", "new feature", or "plan X".
---

When invoked:

1. Read `PRPs/templates/prp_base.md` for the structure.
2. Read `PRPs/examples/prp_auth.md` for tone, depth, and how detailed each section should be.
3. Read `memory-bank/architecture.md`, `memory-bank/database-schema.md`, `memory-bank/api-contracts.md` for current-state context.
4. Ask the user 2–3 sharp clarifying questions (don't fish — make them count).
5. Write the PRP to `PRPs/<kebab-name>.md` following the template exactly. Fill every section. Include:
   - Goal (1 paragraph)
   - Context (load-bearing facts from memory-bank, citing files)
   - File structure (exact paths to create/modify)
   - Task breakdown (numbered, ordered, dependency-aware)
   - Pseudocode for the trickiest function
   - Validation plan (lint + tests + manual smoke steps)
6. Print the path. Do not implement.
```

## Troubleshooting

- **Backend won't start, "connection refused"** — Postgres health check hasn't passed yet. `make logs` and wait for `database system is ready to accept connections`.
- **Frontend can't reach backend** — check `VITE_API_URL` in `.env` matches `http://localhost:8080` for local, or your proxy in prod.
- **Token always rejected** — bumped `TABLE_TOKEN_SECRET` invalidates every existing QR. Owner must regenerate per-table.
- **`gen_random_uuid()` errors** — enable `pgcrypto` (already in `001_init.sql`; check it ran).

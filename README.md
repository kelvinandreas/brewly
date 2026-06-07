# Brewly

> Self-hosted, open-source POS for a single small cafe. No SaaS fees. No vendor lock-in. One `docker compose up`.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.25](https://img.shields.io/badge/Go-1.25+-informational)](https://go.dev)
[![React 19](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://docs.docker.com/compose)

⚠️ Status: Brewly is currently an experiment and is not production-ready. It was built primarily to explore AI-assisted ("vibe engineering") development without manually writing application code. For the full story, lessons learned, limitations, and what still needs improvement, see docs/vibe-engineering.md.

**→ [Why I built this in a few hours using vibe engineering](docs/vibe-engineering.md)**

---

## What Brewly does

| Role         | What they get                                                              |
| ------------ | -------------------------------------------------------------------------- |
| **Customer** | Scan QR on table → browse menu → place order → request songs via YouTube   |
| **Cashier**  | POS terminal — pick table, build cart, record payment (cash / QRIS / card) |
| **Kitchen**  | Live KDS board — orders arrive in real time, one tap to advance status     |
| **Owner**    | Menu + table management, staff accounts, song queue, revenue reports       |

No customer accounts. No payment gateway fees. Works offline from the internet (except YouTube search).

---

## Quick start

**Requirements:** Docker 25+ · Docker Compose v2 · `make`

### 1. Clone and configure

```bash
git clone https://github.com/kelvinandreas/brewly
cd brewly
cp .env.example .env
```

Open `.env` and set the following:

```dotenv
POSTGRES_PASSWORD=pick_a_strong_password

# Generate each secret with: openssl rand -hex 32
JWT_SECRET=...           # staff access tokens (15 min)
REFRESH_SECRET=...       # httpOnly refresh cookies (7 days)
TABLE_TOKEN_SECRET=...   # per-table customer tokens (4 hours)

# Optional — song search is disabled without this key
# Get one free: console.cloud.google.com → YouTube Data API v3
YOUTUBE_API_KEY=...
```

### 2. Start everything

```bash
make dev
make migrate
```

This brings up PostgreSQL, the Go backend (hot reload via Air), and the Vite dev server. Migrations run automatically on first boot.

| Service         | URL                           |
| --------------- | ----------------------------- |
| Staff dashboard | http://localhost:5173         |
| Backend API     | http://localhost:8080         |
| Health check    | http://localhost:8080/healthz |

### 3. Register the owner account

Open http://localhost:5173 — the login page shows a **Register owner** form on first visit. Fill it in. That account has full access.

---

## Cafe owner setup

After registering, the dashboard is empty — that's expected. Complete these steps once before opening day.

### Step 1 — Build your menu

**Dashboard → Menu**

1. **Add Category** — e.g. _Coffee_, _Food_, _Cold Drinks_
2. **Add Item** inside each category — name, price, optional description
3. Toggle **Available / Unavailable** anytime (e.g. when something sells out)

### Step 2 — Set up your tables

**Dashboard → Tables**

1. **Add Table** — enter a label like _Table 1_ or _Bar Seat A_
2. Repeat for every seat in your cafe
3. Click **QR** next to a table → download the PNG → print and laminate it

> Lost a QR or worried it leaked? Click **Regenerate** on that table — the old code is invalidated instantly.

### Step 3 — Create staff accounts

**Dashboard → Staff** _(owner-only)_

- Add a **cashier** account for counter staff
- Add a **kitchen** account for the kitchen display tablet

Staff log in at the same URL with their own credentials.

### Step 4 — Daily workflow

| Who     | Page          | Job                                      |
| ------- | ------------- | ---------------------------------------- |
| Cashier | `/cashier`    | Build orders, record payments            |
| Kitchen | `/kitchen`    | Watch live orders, tap to advance status |
| Owner   | `/song-queue` | DJ board — mark playing, skip            |
| Owner   | `/reports`    | Revenue, best sellers, hourly volume     |

Customers scan their table QR and order directly — no app download, no account.

## Development reference

```bash
# Stack
make dev              # start postgres + backend (Air) + frontend (Vite)
make dev-down         # stop everything

# Database
make migrate                          # apply pending migrations
make migrate-new name=add_loyalty     # scaffold next migration file
make psql                             # psql shell into dev DB

# Quality
make test             # go test ./... + vitest
make test-backend     # backend only
make test-frontend    # frontend only
make lint             # golangci-lint + eslint + prettier --check
make fmt              # gofmt + prettier --write

# Utilities
make logs             # tail docker compose logs
make commit           # interactive conventional-commit helper
```

---

## Project structure

```
brewly/
├── backend/
│   ├── cmd/api/main.go       # entry point — dependency wiring
│   ├── internal/
│   │   ├── domain/           # entities, repo interfaces, errors, constants
│   │   ├── usecase/          # business logic (no HTTP, no DB)
│   │   ├── repository/       # GORM — the only place it appears
│   │   ├── handler/          # parse request → call usecase → respond
│   │   └── middleware/       # JWT auth, table-token auth, CORS, recovery
│   ├── migrations/           # forward-only numbered SQL files
│   └── pkg/                  # jwt, sse, qrcode, youtube, response helpers
│
└── frontend/src/
    ├── routes/               # one file per page; _auth.* routes require login
    ├── hooks/                # all TanStack Query fetching and mutations
    ├── lib/                  # api.ts (fetch), auth.ts (token store), currency.ts
    └── types/api.ts          # TypeScript mirrors of Go response DTOs
```

---

## Architecture

Clean Architecture — `handler → usecase → domain ← repository`. Each layer depends only on the one above it via interfaces; nothing crosses. GORM is confined to `repository/`. Business rules are confined to `usecase/`. Real-time updates use a generic SSE fan-out broker — no WebSockets, no Redis, no message queue.

Full rationale in [docs/ADR/](docs/ADR/).

---

## Tech stack

|             |                                                                            |
| ----------- | -------------------------------------------------------------------------- |
| Backend     | Go 1.25, Chi v5, GORM, PostgreSQL 16                                       |
| Frontend    | React 19, Vite, TanStack Router + Query, Tailwind CSS                      |
| Auth        | golang-jwt/jwt v5 — staff 15 min access / 7 day refresh / 4 hr table token |
| Real-time   | SSE fan-out broker (`pkg/sse`)                                             |
| QR codes    | `skip2/go-qrcode` — server-side PNG                                        |
| Song search | YouTube Data API v3 (optional, degrades gracefully)                        |
| Deploy      | Docker Compose — dev and prod variants                                     |

---

## Docs

|                                                      |                                                              |
| ---------------------------------------------------- | ------------------------------------------------------------ |
| [docs/vibe-engineering.md](docs/vibe-engineering.md) | How this was built with AI, the PRP workflow, prompting tips |
| [docs/ENGINEERING.md](docs/ENGINEERING.md)           | Full local setup, Makefile reference, migration workflow     |
| [docs/specification.md](docs/specification.md)       | User stories, feature specs, security model, non-goals       |
| [docs/ADR/](docs/ADR/)                               | Six architecture decisions — why Chi, clean arch, GORM, SSE  |
| [memory-bank/](memory-bank/)                         | Living docs: schema, API contracts, architecture diagram     |

---

## License

[MIT](LICENSE)

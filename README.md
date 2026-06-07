# Brewly

> An open-source, self-hosted POS for the cafe across the street.

<!-- Screenshots: ./docs/screenshots/dashboard.png, kds.png, table-flow.png — placeholders for now, real shots come later -->

## Features

- Order management with cashier POS terminal + table QR ordering
- Kitchen Display System with live SSE updates
- Customer song request via YouTube Data API v3 (per-table rate limited)
- Manual payments (cash / QRIS / card) — no gateway lock-in
- Daily / weekly / monthly revenue + best-seller reports
- Anonymous customer flow — no customer accounts, signed table tokens, regeneratable per table
- One-command Docker Compose deploy

## Quick start

```bash
git clone https://github.com/your-handle/brewly
cd brewly
cp .env.example .env
# edit .env — set DB_PASSWORD, JWT_SECRET, TABLE_TOKEN_SECRET, YOUTUBE_API_KEY
make dev
```

The dashboard is at http://localhost:5173, API at http://localhost:8080.
On first start, register the owner account at `/login`.

## Tech stack

`Go 1.23` · `Chi` · `GORM` · `PostgreSQL 16` · `React 19` · `TanStack Router/Query` · `Tailwind CSS` · `Docker Compose`

## Documentation

- [Engineering guide](docs/ENGINEERING.md) — local setup, Makefile reference, migration workflow
- [Full specification](docs/specification.md) — user stories, feature specs, security model
- [Architecture decisions](docs/ADR/) — why we chose what we chose

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). PRs welcome — please use conventional commits and `make lint && make test` before pushing.

## License

[MIT](LICENSE)

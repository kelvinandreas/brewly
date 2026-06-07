# Tech stack

Pinned versions for reproducibility. Bump intentionally, never opportunistically.

## Backend

| Package                                | Version  | Purpose                     |
| -------------------------------------- | -------- | --------------------------- |
| go                                     | 1.25.x   | Language runtime            |
| github.com/go-chi/chi/v5               | v5.1.0   | HTTP router                 |
| github.com/go-chi/cors                 | v1.2.x   | CORS middleware             |
| gorm.io/gorm                           | v1.25.x  | ORM                         |
| gorm.io/driver/postgres                | v1.5.x   | Postgres dialect            |
| github.com/golang-jwt/jwt/v5           | v5.x     | JWT                         |
| github.com/google/uuid                 | v1.6.x   | UUID v4                     |
| github.com/go-playground/validator/v10 | v10.x    | Request validation          |
| github.com/skip2/go-qrcode             | v0.0.0-… | QR PNG generation           |
| github.com/joho/godotenv               | v1.5.x   | `.env` loading (dev only)   |
| github.com/rs/zerolog                  | v1.x     | Structured logs             |
| github.com/stretchr/testify            | v1.x     | Test assertions             |
| github.com/air-verse/air               | v1.x     | Hot reload (dev image only) |

## Frontend

| Package                | Version | Purpose             |
| ---------------------- | ------- | ------------------- |
| react                  | ^19.0.0 | UI                  |
| react-dom              | ^19.0.0 | DOM renderer        |
| @tanstack/react-router | ^1.0.0  | File-based routing  |
| @tanstack/react-query  | ^5.0.0  | Data fetching       |
| react-hook-form        | ^7.x    | Forms               |
| zod                    | ^3.x    | Schema validation   |
| tailwindcss            | ^3.x    | Styling             |
| @radix-ui/react-\*     | latest  | Headless primitives |
| lucide-react           | latest  | Icons               |
| vite                   | ^5.x    | Bundler             |
| typescript             | ^5.x    | Type system         |
| vitest                 | ^1.x    | Unit tests          |
| @testing-library/react | ^16.x   | Component tests     |

## Infra

| Tool           | Version     | Purpose                                |
| -------------- | ----------- | -------------------------------------- |
| postgres       | 16-alpine   | Database                               |
| nginx          | 1.27-alpine | Frontend static + reverse proxy (prod) |
| docker         | 25+         | Containerization                       |
| docker compose | v2          | Orchestration                          |

## Upgrade notes

- **React 19** is mandatory — TanStack Router/Query v5 paths assume it. Don't downgrade.
- **GORM 1.25.10+** ships native pgx driver compatibility; required for `gen_random_uuid()` works without extension overhead.
- **Postgres 16** introduces `gen_random_uuid()` in `pgcrypto` extension by default. Enable in `001_init.sql`.

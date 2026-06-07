.PHONY: dev dev-down migrate migrate-new test test-backend test-frontend lint fmt commit build-prod logs psql seed

DOCKER_COMPOSE := docker compose

dev:
	$(DOCKER_COMPOSE) up

dev-down:
	$(DOCKER_COMPOSE) down

migrate:
	bash scripts/migrate.sh

migrate-new:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-new name=add_foo"; exit 1; fi
	@bash scripts/new_migration.sh "$(name)"

test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && pnpm test

lint:
	cd backend && golangci-lint run ./...
	cd frontend && pnpm lint && pnpm format:check

fmt:
	cd backend && gofmt -w .
	cd frontend && pnpm format

commit:
	bash scripts/commit.sh

build-prod:
	$(DOCKER_COMPOSE) -f docker-compose.prod.yml build

logs:
	$(DOCKER_COMPOSE) logs -f

psql:
	$(DOCKER_COMPOSE) exec postgres psql -U $${POSTGRES_USER:-brewly} -d $${POSTGRES_DB:-brewly}

seed:
	@echo "seed: create scripts/seed.sh with your fixture data first"

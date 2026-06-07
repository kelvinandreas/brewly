#!/usr/bin/env bash
set -euo pipefail

MIGRATIONS_DIR="backend/migrations"

if [ ! -d "$MIGRATIONS_DIR" ] || [ -z "$(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null)" ]; then
  echo "No SQL migrations found in $MIGRATIONS_DIR"
  exit 0
fi

for f in $(ls "$MIGRATIONS_DIR"/*.sql | sort); do
  echo "→ Applying $f"
  docker compose exec -T postgres psql \
    -U "${POSTGRES_USER:-brewly}" \
    -d "${POSTGRES_DB:-brewly}" \
    < "$f"
done

echo "✓ Migrations applied"

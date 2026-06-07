#!/usr/bin/env bash
set -euo pipefail

NAME="${1:-}"
if [[ -z "$NAME" ]]; then
  echo "Usage: $0 <migration_name>" >&2
  exit 1
fi

MIGRATIONS_DIR="backend/migrations"
COUNT=$(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | wc -l)
NEXT=$(printf "%03d" $((COUNT + 1)))
FILE="$MIGRATIONS_DIR/${NEXT}_${NAME}.sql"

cat > "$FILE" <<SQL
-- Migration: ${NEXT}_${NAME}
-- Created: $(date -u +"%Y-%m-%dT%H:%M:%SZ")

SQL

echo "Created: $FILE"

#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_HOST_PORT="${DB_HOST_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-1234}"
DB_NAME="${DB_NAME:-spice_ledger}"
MODE="local"

for arg in "$@"; do
  case "$arg" in
    local|docker) MODE="$arg" ;;
  esac
done

# When using Docker MySQL from the host, connect via mapped port
if [[ "$MODE" == "local" && "$DB_HOST" == "db" ]]; then
  DB_HOST="127.0.0.1"
  DB_PORT="$DB_HOST_PORT"
fi

mysql_exec() {
  if [[ "$MODE" == "docker" ]]; then
    docker compose --profile infra exec -T db mysql -uroot -p"${DB_PASSWORD}" "$@"
  else
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"${DB_PASSWORD}" "$@"
  fi
}

echo "Waiting for MySQL at ${DB_HOST}:${DB_PORT}..."
for i in {1..30}; do
  if mysql_exec -e "SELECT 1" >/dev/null 2>&1; then
    break
  fi
  if [[ "$i" -eq 30 ]]; then
    echo "MySQL is not reachable. Start it with: brew services start mysql  OR  make up-db"
    exit 1
  fi
  sleep 1
done

mysql_exec -e "CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\`;"

echo "Applying versioned migrations..."
if [[ "$MODE" == "docker" || "$DB_HOST" == "127.0.0.1" && "$DB_PORT" == "$DB_HOST_PORT" ]]; then
  docker compose --profile infra run --rm migrate up
else
  go run ./cmd/migrate/main.go up
fi

echo "Database '${DB_NAME}' is ready. View versions: make migrate-status"

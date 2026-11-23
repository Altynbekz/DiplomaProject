#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
python3 scripts/setup-local.py
set -a
. ./.env
set +a
mkdir -p artifacts/local-db
cluster="$root/artifacts/local-db/data"
port=${LOCAL_DB_PORT:-55440}
if [ ! -s "$cluster/PG_VERSION" ]; then
 initdb -D "$cluster" -A scram-sha-256 -U diploma_migrate --pwfile="$root/.secrets/db_password" --no-locale -E UTF8
fi
if ! pg_ctl -D "$cluster" status >/dev/null 2>&1; then
 pg_ctl -D "$cluster" -l "$root/artifacts/local-db/postgres.log" -o "-h 127.0.0.1 -p $port -k $root/artifacts/local-db" -w start
fi
export PGPASSWORD="$(cat .secrets/db_password)"
if ! psql -h 127.0.0.1 -p "$port" -U diploma_migrate -d postgres -Atc "SELECT datname FROM pg_database WHERE datname='diploma'" | rg -q '^diploma$'; then
 createdb -h 127.0.0.1 -p "$port" -U diploma_migrate diploma
fi
export MIGRATION_DATABASE_URL="postgres://diploma_migrate:$PGPASSWORD@127.0.0.1:$port/diploma?sslmode=disable"
goose -dir backend/db/migrations postgres "$MIGRATION_DATABASE_URL" up
runtime_password=$(cat .secrets/runtime_password)
DIPLOMA_RUNTIME_PASSWORD="$runtime_password" psql "$MIGRATION_DATABASE_URL" -f scripts/runtime-login.sql >/dev/null
export DATABASE_URL="postgres://diploma_runtime:$runtime_password@127.0.0.1:$port/diploma?sslmode=disable"
export APP_API_TOKEN="$(cat .secrets/api_token)"
export APP_OPERATOR_ID=${APP_OPERATOR_ID:-local-operator}
export JUICE_SHOP_ORIGIN=${JUICE_SHOP_ORIGIN:-http://juice-shop:3000}
(cd backend && go build -o ../artifacts/diploma-api ./cmd/api)
./artifacts/diploma-api & api_pid=$!
(cd frontend && exec ./node_modules/.bin/vite --host 127.0.0.1 --port 5173) & ui_pid=$!
cleanup(){ kill "$api_pid" "$ui_pid" 2>/dev/null || true; }
trap cleanup EXIT HUP INT TERM
printf '%s\n' 'Workspace: http://127.0.0.1:5173. Token: .secrets/api_token. PostgreSQL retained; no scans.'
wait

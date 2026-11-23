#!/bin/sh
set -eu
if [ "${1:-api}" = migrate ]; then
 export GOOSE_DBSTRING="postgres://diploma_migrate:$(cat /run/secrets/db_password)@db:5432/diploma?sslmode=disable"
 export GOOSE_DRIVER=postgres
 exec goose -dir /app/migrations up
fi
export DATABASE_URL="postgres://diploma_runtime:$(cat /run/secrets/runtime_password)@db:5432/diploma?sslmode=disable"
export APP_API_TOKEN="$(cat /run/secrets/api_token)"
exec /app/api

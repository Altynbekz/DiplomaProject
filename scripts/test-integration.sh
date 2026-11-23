#!/bin/sh
set -eu
# Own a fresh temporary cluster; never reset an existing/user-supplied database.
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
for cmd in initdb pg_ctl createdb goose go; do command -v "$cmd" >/dev/null || { echo "Missing host test dependency: $cmd" >&2; exit 1; }; done
cluster=$(mktemp -d "${TMPDIR:-/tmp}/diploma-test.XXXXXX")
port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')
cleanup() { pg_ctl -D "$cluster/data" -m fast -w stop >/dev/null 2>&1 || true; rm -rf -- "$cluster"; }
trap cleanup EXIT HUP INT TERM
initdb -D "$cluster/data" -A trust -U diploma_migrate --no-locale -E UTF8 >/dev/null
pg_ctl -D "$cluster/data" -l "$cluster/postgres.log" -o "-h 127.0.0.1 -p $port -k $cluster" -w start >/dev/null
createdb -h 127.0.0.1 -p "$port" -U diploma_migrate diploma_test
export TEST_DATABASE_URL="postgres://diploma_migrate@127.0.0.1:$port/diploma_test?sslmode=disable"
# Exercise empty database, previous revision upgrade, and no-op upgrade.
goose -dir "$root/backend/db/migrations" postgres "$TEST_DATABASE_URL" up-to 1
goose -dir "$root/backend/db/migrations" postgres "$TEST_DATABASE_URL" up-to 4
goose -dir "$root/backend/db/migrations" postgres "$TEST_DATABASE_URL" up
goose -dir "$root/backend/db/migrations" postgres "$TEST_DATABASE_URL" up
createdb -h 127.0.0.1 -p "$port" -U diploma_migrate diploma_genuine
export TEST_GENUINE_DATABASE_URL="postgres://diploma_migrate@127.0.0.1:$port/diploma_genuine?sslmode=disable"
goose -dir "$root/backend/db/migrations" postgres "$TEST_GENUINE_DATABASE_URL" up
cd "$root/backend"
go test -race -count=1 -v ./internal/httpapi -run TestIntegration

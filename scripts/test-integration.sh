#!/bin/sh
set -eu
# Own a fresh temporary cluster; never reset an existing/user-supplied database.
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
for cmd in initdb pg_ctl createdb psql goose go; do command -v "$cmd" >/dev/null || { echo "Missing host test dependency: $cmd" >&2; exit 1; }; done
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
goose -dir "$root/backend/db/migrations" postgres "$TEST_DATABASE_URL" up-to 6
goose -dir "$root/backend/db/migrations" postgres "$TEST_DATABASE_URL" up
goose -dir "$root/backend/db/migrations" postgres "$TEST_DATABASE_URL" up
createdb -h 127.0.0.1 -p "$port" -U diploma_migrate diploma_genuine
export TEST_GENUINE_DATABASE_URL="postgres://diploma_migrate@127.0.0.1:$port/diploma_genuine?sslmode=disable"
goose -dir "$root/backend/db/migrations" postgres "$TEST_GENUINE_DATABASE_URL" up
# Seed a populated v7 database and compare every existing row through the upgrade.
createdb -h 127.0.0.1 -p "$port" -U diploma_migrate diploma_upgrade
upgrade="postgres://diploma_migrate@127.0.0.1:$port/diploma_upgrade?sslmode=disable"
goose -dir "$root/backend/db/migrations" postgres "$upgrade" up-to 7
psql "$upgrade" -X -v ON_ERROR_STOP=1 -q -f "$root/backend/testdata/cvss/upgrade-v7.sql"
snapshot_rows() {
    psql "$upgrade" -X -v ON_ERROR_STOP=1 -Atc 'SELECT jsonb_build_array(
      (SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM targets t),
      (SELECT jsonb_agg(to_jsonb(r) ORDER BY id) FROM report_imports r),
      (SELECT jsonb_agg(to_jsonb(f) ORDER BY id) FROM findings f),
      (SELECT jsonb_agg(to_jsonb(o) ORDER BY id) FROM observations o),
      (SELECT jsonb_agg(to_jsonb(v) ORDER BY id) FROM verification_reviews v),
      (SELECT jsonb_agg(to_jsonb(c) ORDER BY id) FROM classification_revisions c),
      (SELECT jsonb_agg(to_jsonb(r) ORDER BY id) FROM context_revisions r))'
}
snapshot_rows > "$cluster/before.json"
goose -dir "$root/backend/db/migrations" postgres "$upgrade" up
snapshot_rows > "$cluster/after.json"
cmp "$cluster/before.json" "$cluster/after.json"
test "$(psql "$upgrade" -X -v ON_ERROR_STOP=1 -Atc 'SELECT count(*) FROM assessment_snapshots')" = 0
printf '%s\n' 'Populated v7 upgrade preserved all M2 rows and initialized empty assessment storage.'
cd "$root/backend"
go test -race -count=1 -v ./internal/httpapi -run TestIntegration

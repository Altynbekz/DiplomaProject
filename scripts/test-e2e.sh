#!/bin/sh
set -eu
if [ "$#" -eq 0 ]; then
    "$0" synthetic
    "$0" genuine
    exit
fi
case "$1" in
    synthetic) upload_case='synthetic report'; restart_case='persisted rows survive' ;;
    genuine) upload_case='genuine report survives reload'; restart_case='genuine findings survive restart' ;;
    *) echo "Expected synthetic or genuine suite" >&2; exit 2 ;;
esac
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cluster=$(mktemp -d "${TMPDIR:-/tmp}/diploma-e2e.XXXXXX")
ports=$(python3 -c 'import socket; a=[]; ss=[]
for _ in range(3):
 s=socket.socket(); s.bind(("127.0.0.1",0)); a.append(str(s.getsockname()[1])); ss.append(s)
print(" ".join(a))')
set -- $ports
db_port=$1; api_port=$2; ui_port=$3
api_pid=; ui_pid=
cleanup(){ [ -z "$ui_pid" ] || kill "$ui_pid" 2>/dev/null || true; [ -z "$api_pid" ] || kill "$api_pid" 2>/dev/null || true; pg_ctl -D "$cluster/data" -m fast -w stop >/dev/null 2>&1 || true; rm -rf -- "$cluster"; }
trap cleanup EXIT HUP INT TERM
initdb -D "$cluster/data" -A trust -U diploma_migrate --no-locale -E UTF8 >/dev/null
pg_ctl -D "$cluster/data" -l "$cluster/postgres.log" -o "-h 127.0.0.1 -p $db_port -k $cluster" -w start >/dev/null
createdb -h 127.0.0.1 -p "$db_port" -U diploma_migrate diploma_test
migration_url="postgres://diploma_migrate@127.0.0.1:$db_port/diploma_test?sslmode=disable"
goose -dir "$root/backend/db/migrations" postgres "$migration_url" up
psql "$migration_url" -c 'ALTER ROLE diploma_runtime LOGIN' >/dev/null
export APP_API_TOKEN=$(python3 -c 'import secrets;print(secrets.token_hex(32))')
export E2E_API_TOKEN=$APP_API_TOKEN
export DATABASE_URL="postgres://diploma_runtime@127.0.0.1:$db_port/diploma_test?sslmode=disable"
export APP_OPERATOR_ID=e2e-operator
export APP_LISTEN="127.0.0.1:$api_port"
export API_PROXY_URL="http://127.0.0.1:$api_port"
export E2E_BASE_URL="http://127.0.0.1:$ui_port"
(cd "$root/backend" && go build -o "$cluster/api" ./cmd/api)
"$cluster/api" >"$cluster/api.log" 2>&1 & api_pid=$!
(cd "$root/frontend" && exec ./node_modules/.bin/vite --host 127.0.0.1 --port "$ui_port") >"$cluster/ui.log" 2>&1 & ui_pid=$!
python3 - "$api_port" "$ui_port" <<'WAIT'
import sys,time,urllib.request
for port,path in [(sys.argv[1],'/readyz'),(sys.argv[2],'/')]:
 for attempt in range(100):
  try:
   with urllib.request.urlopen('http://127.0.0.1:'+port+path,timeout=1) as r:
    if r.status==200:break
  except OSError:time.sleep(.1)
 else:raise SystemExit('Test stack did not become ready')
WAIT
(cd "$root/frontend" && npm run test:e2e -- --grep "$upload_case")
# Prove database outage semantics while the API process remains live.
pg_ctl -D "$cluster/data" -m fast -w stop >/dev/null
python3 - "$api_port" <<'OUTAGE'
import sys,urllib.request,urllib.error
base='http://127.0.0.1:'+sys.argv[1]
for path,want in [('/healthz',200),('/readyz',503),('/api/v1/targets',401)]:
 try:
  with urllib.request.urlopen(base+path,timeout=5) as r:code=r.status
 except urllib.error.HTTPError as e:code=e.code
 if code!=want:raise SystemExit(f'{path}: expected {want}, got {code}')
print('DB outage: health=200, readiness=503, missing bearer=401')
OUTAGE
kill "$api_pid"
wait "$api_pid" || true
pg_ctl -D "$cluster/data" -l "$cluster/postgres.log" -o "-h 127.0.0.1 -p $db_port -k $cluster" -w start >/dev/null
"$cluster/api" >"$cluster/api.log" 2>&1 & api_pid=$!
python3 - "$api_port" <<'RESTART'
import sys,time,urllib.request
for attempt in range(100):
 try:
  with urllib.request.urlopen('http://127.0.0.1:'+sys.argv[1]+'/readyz',timeout=1) as r:
   if r.status==200:break
 except OSError:time.sleep(.1)
else:raise SystemExit('Restart failed readiness')
RESTART
(cd "$root/frontend" && npm run test:e2e -- --grep "$restart_case")

#!/bin/sh
set -eu
DIPLOMA_RUNTIME_PASSWORD="$(cat /run/secrets/runtime_password)" psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -v ON_ERROR_STOP=1 <<'SQL'
\getenv runtime_password DIPLOMA_RUNTIME_PASSWORD
CREATE ROLE diploma_runtime LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT PASSWORD :'runtime_password';
SQL

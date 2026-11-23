#!/bin/sh
set -eu

# Build the same PostgreSQL release used by the application without relying on
# the changing package versions installed in a hosted runner image.
version=15.19
digest=e1a64a87a46b825b88c082e4518161a47aab53c45694964f8ba1df28f7859f89
prefix=${1:?Usage: install-test-postgres.sh /absolute/new/prefix}
case "$prefix" in
  /*) ;;
  *) printf '%s\n' 'PostgreSQL prefix must be an absolute path.' >&2; exit 1 ;;
esac
if [ -e "$prefix" ]; then
  printf '%s\n' 'Choose a new PostgreSQL installation prefix.' >&2
  exit 1
fi
if [ "$(id -u)" -eq 0 ]; then
  printf '%s\n' 'Run the test PostgreSQL installation as a non-root user.' >&2
  exit 1
fi

source_dir=$(mktemp -d "${TMPDIR:-/tmp}/diploma-pg-source.XXXXXX")
trap 'rm -rf -- "$source_dir"' EXIT
trap 'exit 1' HUP INT TERM
cd "$source_dir"
archive="postgresql-$version.tar.bz2"
curl --fail --silent --show-error --location --proto '=https' --retry 3 \
  "https://ftp.postgresql.org/pub/source/v$version/$archive" --output "$archive"
python3 - "$archive" "$digest" <<'PY'
import hashlib
import sys
with open(sys.argv[1], "rb") as archive:
    actual = hashlib.file_digest(archive, "sha256").hexdigest()
if actual != sys.argv[2]:
    raise SystemExit("PostgreSQL source checksum does not match the pinned release")
PY
tar -xjf "$archive"
cd "postgresql-$version"
if ! (
  ./configure --prefix="$prefix" --without-readline --without-zlib
  make -s -j "${POSTGRES_BUILD_JOBS:-4}"
  make -s install
) >"$source_dir/build.log" 2>&1; then
  tail -80 "$source_dir/build.log" >&2
  exit 1
fi
"$prefix/bin/initdb" --version

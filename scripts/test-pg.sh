#!/usr/bin/env bash
# Runs the PostgreSQL integration tests (every internal/*/postgres package, plus the cmd/ packages listed below)
# against a disposable server. Each test creates its own database and drops it afterwards.
#
# - OPENFIEF_PG_DSN set: use that server. It must allow CREATE DATABASE; never point it at real data.
# - Otherwise: start a throwaway cluster from the local PostgreSQL binaries (initdb and pg_ctl on PATH, or the newest
#   /usr/lib/postgresql/*/bin), listening only on a Unix socket in a temporary directory, and remove it on exit.
#   As root, the cluster runs as the postgres user, because PostgreSQL refuses to run as root.
# Extra arguments go to go test, e.g. task test:pg -- -run TestAppend -v
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
# cmd/demo-seed loads the demo dataset end to end into a fresh database; cmd/core serves it over the read API.
mapfile -t pkgs < <(go list ./internal/... | grep '/postgres$'; go list ./cmd/demo-seed ./cmd/core)
if [ "${#pkgs[@]}" -eq 0 ]; then
  echo "test:pg: no postgres packages"
  exit 0
fi

if [ -n "${OPENFIEF_PG_DSN:-}" ]; then
  echo "test:pg: using OPENFIEF_PG_DSN"
  exec go test -count=1 "$@" "${pkgs[@]}"
fi

bindir=""
if ! command -v initdb >/dev/null 2>&1; then
  bindir="$(find /usr/lib/postgresql -maxdepth 2 -name bin -type d 2>/dev/null | sort -V | tail -n1)"
  if [ -z "$bindir" ]; then
    echo "test:pg: no PostgreSQL binaries found; install PostgreSQL or set OPENFIEF_PG_DSN" >&2
    exit 1
  fi
  bindir="$bindir/"
fi

dir="$(mktemp -d)"
as_pg=()
if [ "$(id -u)" -eq 0 ]; then
  chown postgres "$dir"
  as_pg=(runuser -u postgres --)
fi
trap '"${as_pg[@]}" "${bindir}pg_ctl" -D "$dir/data" -m immediate stop >/dev/null 2>&1 || true; rm -rf "$dir"' EXIT

"${as_pg[@]}" "${bindir}initdb" -D "$dir/data" -U openfief --auth=trust --no-sync -E UTF8 --locale=C >/dev/null
"${as_pg[@]}" "${bindir}pg_ctl" -D "$dir/data" -l "$dir/server.log" -w \
  -o "-k $dir -c listen_addresses='' -c fsync=off -c synchronous_commit=off -c full_page_writes=off" start >/dev/null
echo "test:pg: started $("${bindir}postgres" --version) in $dir"

OPENFIEF_PG_DSN="host=$dir user=openfief dbname=postgres sslmode=disable" go test -count=1 "$@" "${pkgs[@]}"

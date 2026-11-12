#!/usr/bin/env bash
# Runs the authorization contract test (internal/identity/spicedb, build tag spicedb) against a throwaway SpiceDB
# in testing mode: an in-memory datastore per preshared key, plaintext gRPC on a loopback port, nothing persisted.
# The test checks that SpiceDB, with authz/schema.zed, answers every holder's actions as identity.MayAct does (D-164).
#
# SPICEDB names the spicedb binary (task test:authz builds the pinned one into .bin/); it is on PATH otherwise.
# Extra arguments go to go test, e.g. task test:authz -- -run TestContract -v
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
spicedb="${SPICEDB:-spicedb}"
if ! command -v "$spicedb" >/dev/null 2>&1 && [ ! -x "$spicedb" ]; then
  echo "test:authz: no spicedb binary; run task test:authz, which builds the pinned one" >&2
  exit 1
fi

# A free loopback port: bind to 0, read it back, release it.
port="$(python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')"
log="$(mktemp)"
"$spicedb" serve-testing \
  --grpc-addr "127.0.0.1:$port" \
  --http-enabled=false --readonly-grpc-enabled=false --readonly-http-enabled=false \
  >"$log" 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; rm -f "$log"' EXIT

# Wait until it listens, for at most 30 seconds, and say so if it died.
for _ in $(seq 1 300); do
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "test:authz: spicedb stopped at start:" >&2
    cat "$log" >&2
    exit 1
  fi
  if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then
    break
  fi
  sleep 0.1
done

OPENFIEF_SPICEDB_TEST_ENDPOINT="127.0.0.1:$port" go test -count=1 -tags spicedb "$@" ./internal/identity/spicedb/

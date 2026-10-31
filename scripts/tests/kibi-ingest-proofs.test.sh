#!/usr/bin/env bash
# Tests for scripts/kibi-ingest-proofs.sh with a fake kibi that logs each ingest-proof input and answers success
# or a refusal; no real knowledge base is touched (kibi: TEST-tooling-kibi-ingest-proofs).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# The fake kibi appends its route, its flag and the --input file's JSON on one line to $FAKE_LOG; it refuses the input whose
# snapshot is $FAKE_REFUSE.
cat >"$work/kibi" <<'EOF'
#!/usr/bin/env bash
input="$(gojq -c . "$3")"
echo "$1 $2 $input" >>"$FAKE_LOG"
if [ "$(gojq -r .snapshot <<<"$input")" = "${FAKE_REFUSE:-}" ]; then
  echo '{"status":"error","error":{"code":"SNAPSHOT_MISMATCH"}}'
  exit 1
fi
echo '{"status":"success","data":{"results":[{"testId":"TEST-e2e-x","outcome":"passed"}]}}'
EOF
chmod +x "$work/kibi"

# artifact writes a proof-run artifact for one cluster: its snapshot and when its run finished.
artifact() {
  local cluster="$1" snapshot="$2" finished="$3"
  mkdir -p "$work/runs/e2e-proof-$cluster"
  printf '{"version":"kibi.proof-run.v1","integration":"e2e","code_snapshot":"%s","run":{"outcome":"passed","finished_at":"%s"}}\n' \
    "$snapshot" "$finished" >"$work/runs/e2e-proof-$cluster/e2e.json"
}

# ingest runs kibi-ingest-proofs.sh over $work/runs with the fake kibi.
ingest() {
  : >"$work/log"
  FAKE_LOG="$work/log" FAKE_REFUSE="${FAKE_REFUSE:-}" PATH="$work:$PATH" bash "$root/scripts/kibi-ingest-proofs.sh" "$work/kibi" "$work/runs" >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" "$work/log" >&2
    failures=$((failures + 1))
  fi
}

# snapshots prints the snapshots kibi was handed, in call order, on one line.
snapshots() { sed 's/^ingest-proof --input //' "$work/log" | gojq -r .snapshot | paste -sd, -; }

mkdir -p "$work/runs"
ingest && rc=0 || rc=$?
check "no artifact fails without calling kibi" "$([ "$rc" -ne 0 ] && [ ! -s "$work/log" ] && echo yes || echo no)"

artifact k3d bbbb 2026-10-31T12:30:00.000Z
artifact kind aaaa 2026-10-31T12:20:00.000Z
ingest && rc=0 || rc=$?
check "both clusters' runs are ingested, the earlier finished first" \
  "$([ "$rc" -eq 0 ] && [ "$(snapshots)" = aaaa,bbbb ] && echo yes || echo no)"
check "kibi gets each artifact whole, with its own snapshot, through ingest-proof" \
  "$(head -1 "$work/log" | grep -q '^ingest-proof --input {"artifact":{"code_snapshot":"aaaa","integration":"e2e","run":{"finished_at":"2026-10-31T12:20:00.000Z","outcome":"passed"},"version":"kibi.proof-run.v1"},"snapshot":"aaaa"}$' && echo yes || echo no)"

FAKE_REFUSE=aaaa ingest && rc=0 || rc=$?
check "a refused artifact fails the script before the next one" \
  "$([ "$rc" -ne 0 ] && [ "$(snapshots)" = aaaa ] && grep -q SNAPSHOT_MISMATCH "$work/out" && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "kibi-ingest-proofs.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "kibi-ingest-proofs.test.sh: all $cases checks passed"

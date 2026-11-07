#!/usr/bin/env bash
# Tests for scripts/kibi-coverage.sh with a stand-in kibi that answers pages of a fixed list and counts its calls, in a
# throwaway Git repository with a small .kb, kibi patches and installed kibi versions, so no case touches the real
# knowledge base (kibi: TEST-tooling-kibi-coverage).
set -euo pipefail
# A hook exports GIT_INDEX_FILE and friends; the throwaway repository must not inherit them.
unset GIT_DIR GIT_INDEX_FILE GIT_WORK_TREE GIT_PREFIX GIT_OBJECT_DIRECTORY

root="$(git rev-parse --show-toplevel)"
script="$root/scripts/kibi-coverage.sh"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# The stand-in kibi logs each call's arguments to $FAKE_CALLS. `coverage ... --limit L --offset O` answers rows REQ-<i>
# for i in [O, min(O+L, $TOTAL)), with summary.total $TOTAL (plus $FAKE_EXTRA) and meta.syncState $FAKE_SYNC (default
# fresh); with FAKE_IGNORE_OFFSET=1 every page starts at 0. `query` answers an empty list.
cat >"$work/kibi" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "$*" >>"$FAKE_CALLS"
[ "$1" = coverage ] || { echo '[]'; exit 0; }
limit=0 offset=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --limit) limit="$2"; shift ;;
    --offset) offset="$2"; shift ;;
  esac
  shift
done
[ "${FAKE_IGNORE_OFFSET:-}" = 1 ] && offset=0
end=$((offset + limit))
[ "$end" -gt "$TOTAL" ] && end="$TOTAL"
node -e '
const [o, e, total, sync] = process.argv.slice(1);
const rows = [];
for (let i = Number(o); i < Number(e); i++) rows.push({ id: `REQ-${i}` });
console.log(JSON.stringify({ meta: { syncState: sync }, summary: { total: Number(total) }, rows }));
' "$offset" "$end" "$((TOTAL + ${FAKE_EXTRA:-0}))" "${FAKE_SYNC:-fresh}"
EOF
chmod +x "$work/kibi"

# A repository with what the cache key reads: .kb files, both kibi patches and the installed kibi versions.
repo="$work/repo"
mkdir -p "$repo/.kb/requirements" "$repo/patches" "$repo/node_modules/kibi-cli" "$repo/node_modules/kibi-core"
git -C "$repo" init -q
printf '.kb/usage.log\n/node_modules\n/site/.gen/\n' >"$repo/.gitignore"
echo "requirement one" >"$repo/.kb/requirements/REQ-one.md"
echo "cli patch" >"$repo/patches/kibi-cli+2.36.0.patch"
echo "core patch" >"$repo/patches/kibi-core+0.26.1.patch"
echo '{"version": "2.36.0"}' >"$repo/node_modules/kibi-cli/package.json"
echo '{"version": "0.26.1"}' >"$repo/node_modules/kibi-core/package.json"
git -C "$repo" add -A
cache="$repo/site/.gen/cache/kibi-coverage"

# run TOTAL [VAR=value...] runs the script in the repository for TOTAL requirements with the cache settings cleared,
# then the given environment; the report goes to $work/out.json, stderr to $work/err, the kibi calls to $work/calls.
run() {
  local total="$1"
  shift
  : >"$work/calls"
  (cd "$repo" && env -u CI -u GITHUB_ACTIONS -u KIBI_COVERAGE_CACHE -u KIBI_COVERAGE_PAGE \
    TOTAL="$total" FAKE_CALLS="$work/calls" "$@" bash "$script" "$work/kibi" "$work/out.json") 2>"$work/err"
}

# check records a failed expectation with the case name and the run's stderr.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  err | /' "$work/err" >&2
    sed 's/^/  kibi | /' "$work/calls" >&2
    failures=$((failures + 1))
  fi
}

# ids prints the report's requirement ids on one line; coverage_calls counts the coverage queries kibi answered.
ids() { node -p 'JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).rows.map((r) => r.id).join(" ")' "$work/out.json"; }
coverage_calls() { grep -c '^coverage' "$work/calls" || true; }
kibi_calls() { wc -l <"$work/calls" | tr -d ' '; }
has() { grep -qF -- "$1" "$work/err"; }
cached() { find "$cache" -name '*.json' 2>/dev/null | wc -l | tr -d ' '; }

# Paging, with the cache off: every row once, in order, one timed coverage query per page, and no empty last page
# when the rows end on a page boundary (kibi's summary.total says they ended).
for c in "7 3 3" "6 3 2" "2 3 1" "0 3 1"; do
  set -- $c
  want="$(if [ "$1" -gt 0 ]; then seq 0 $(($1 - 1)) | sed 's/^/REQ-/' | paste -sd' '; fi)"
  run "$1" KIBI_COVERAGE_PAGE="$2" KIBI_COVERAGE_CACHE=off
  timed="$(grep -c 'kibi coverage took' "$work/err" || true)"
  check "$1 rows in pages of $2 come out whole and in order, each of the $3 pages timed" \
    "$([ "$(ids)" = "$want" ] && [ "$(coverage_calls)" -eq "$3" ] && [ "$timed" -eq "$3" ] && echo yes || echo no)"
done

run 301 KIBI_COVERAGE_CACHE=off
check "the default page is 300 requirements, after one query that attaches the store" \
  "$([ "$(grep -c -- '--limit 300 ' "$work/calls")" -eq 2 ] && head -1 "$work/calls" | grep -q '^query' && [ "$(node -p 'JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).rows.length' "$work/out.json")" -eq 301 ] && echo yes || echo no)"

rm -f "$work/out.json"
run 7 KIBI_COVERAGE_PAGE=3 KIBI_COVERAGE_CACHE=off FAKE_IGNORE_OFFSET=1 && rc=0 || rc=$?
check "pages that repeat requirements fail and write no report" \
  "$([ "$rc" -ne 0 ] && has "kibi counts 7" && [ ! -e "$work/out.json" ] && echo yes || echo no)"

run 7 KIBI_COVERAGE_PAGE=3 KIBI_COVERAGE_CACHE=off FAKE_EXTRA=2 && rc=0 || rc=$?
check "pages that miss requirements kibi counts fail" "$([ "$rc" -ne 0 ] && has "kibi counts 9" && echo yes || echo no)"

(cd "$repo" && bash "$script" "$work/kibi") 2>"$work/err" && rc=0 || rc=$?
check "a missing argument is a usage error" "$([ "$rc" -eq 2 ] && has usage && echo yes || echo no)"

# Cache.
rm -rf "$cache"
run 5
check "a first run asks kibi and caches the report" \
  "$([ "$(coverage_calls)" -eq 1 ] && has "regenerated and cached the report for" && [ "$(cached)" -eq 1 ] && echo yes || echo no)"
first="$(ids)"

run 5
check "a second run with the same key doesn't call kibi and writes the same report" \
  "$([ "$(kibi_calls)" -eq 0 ] && has "kibi coverage: reused cached report for" && [ "$(ids)" = "$first" ] && echo yes || echo no)"

echo junk >"$repo/.kb/usage.log"
run 5
check "an ignored file under .kb doesn't change the key" "$([ "$(kibi_calls)" -eq 0 ] && echo yes || echo no)"

echo "requirement one, changed" >"$repo/.kb/requirements/REQ-one.md"
run 5
check "a changed .kb file in the working tree, not yet staged, regenerates" \
  "$([ "$(coverage_calls)" -eq 1 ] && has "no cached report for" && echo yes || echo no)"

echo "requirement two" >"$repo/.kb/requirements/REQ-two.md"
run 5
check "a new .kb file Git doesn't track yet regenerates" "$([ "$(coverage_calls)" -eq 1 ] && echo yes || echo no)"

echo "core patch, changed" >"$repo/patches/kibi-core+0.26.1.patch"
run 5
check "a changed kibi patch regenerates, keeping only the three newest reports" \
  "$([ "$(coverage_calls)" -eq 1 ] && [ "$(cached)" -eq 3 ] && echo yes || echo no)"

echo '{"version": "0.27.0"}' >"$repo/node_modules/kibi-core/package.json"
run 5
check "another installed kibi-core version regenerates" "$([ "$(coverage_calls)" -eq 1 ] && echo yes || echo no)"

run 5
check "the regenerated report is then reused" "$([ "$(kibi_calls)" -eq 0 ] && echo yes || echo no)"

for ci in CI=true GITHUB_ACTIONS=true KIBI_COVERAGE_CACHE=off; do
  before="$(ls -t "$cache")"
  run 5 "$ci"
  check "$ci regenerates although the report is cached, and leaves the cache alone" \
    "$([ "$(coverage_calls)" -eq 1 ] && has "cache skipped" && [ "$(ls -t "$cache")" = "$before" ] && echo yes || echo no)"
done

mkdir -p "$repo/.kb/proof/runs/ci"
echo '{}' >"$repo/.kb/proof/runs/ci/run.json"
before="$(ls -t "$cache")"
run 5
check "ingested proof runs regenerate, since they depend on the whole workspace, and leave the cache alone" \
  "$([ "$(coverage_calls)" -eq 1 ] && has "cache skipped (ingested proof runs" && [ "$(ls -t "$cache")" = "$before" ] && echo yes || echo no)"
rm -rf "$repo/.kb/proof/runs"

echo "requirement three" >"$repo/.kb/requirements/REQ-three.md"
run 5 FAKE_SYNC=stale
check "a report from a store that isn't fresh is written but not cached" \
  "$([ "$(ids)" = "$first" ] && has "not cached" && echo yes || echo no)"
run 5
check "so the next run asks kibi again" "$([ "$(coverage_calls)" -eq 1 ] && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "kibi-coverage.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "kibi-coverage.test.sh: all $cases checks passed"

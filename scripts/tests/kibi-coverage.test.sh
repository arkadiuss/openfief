#!/usr/bin/env bash
# Tests for scripts/kibi-coverage.sh with a stand-in kibi that answers pages of a fixed list (kibi: TEST-tooling-kibi-coverage).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0

# fake kibi answers `coverage ... --limit L --offset O` with rows REQ-<i> for i in [O, min(O+L, $TOTAL)).
cat >"$work/kibi" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
limit=0 offset=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --limit) limit="$2"; shift ;;
    --offset) offset="$2"; shift ;;
  esac
  shift
done
end=$((offset + limit))
[ "$end" -gt "$TOTAL" ] && end="$TOTAL"
gojq -n --argjson o "$offset" --argjson e "$end" '{meta: {}, summary: {}, rows: [range($o; $e) | {id: "REQ-\(.)"}]}'
EOF
chmod +x "$work/kibi"

# run TOTAL PAGE prints the ids the script wrote, one per line.
run() {
  TOTAL="$1" KIBI_COVERAGE_PAGE="$2" bash "$root/scripts/kibi-coverage.sh" "$work/kibi" "$work/out.json" >/dev/null
  gojq -r '.rows[].id' "$work/out.json"
}

for case in "7 3" "6 3" "2 3" "0 3"; do
  set -- $case
  want="$(seq 0 $(($1 - 1)) 2>/dev/null | sed 's/^/REQ-/')"
  [ "$1" -eq 0 ] && want=""
  got="$(run "$1" "$2")"
  if [ "$got" != "$want" ]; then
    echo "FAIL: $1 rows in pages of $2: got [$(echo $got)], want [$(echo $want)]" >&2
    failures=$((failures + 1))
  fi
done

if bash "$root/scripts/kibi-coverage.sh" "$work/kibi" 2>/dev/null; then
  echo "FAIL: a missing argument should fail" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "kibi-coverage: 5 cases passed"

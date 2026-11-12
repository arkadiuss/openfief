#!/usr/bin/env bash
# Tests for the attach-time dedupe in patches/kibi-core+0.31.0.patch (docs/dev/kibi.md#setup-notes), run on the
# installed kibi-core with kibi's own SWI-Prolog (kibi: TEST-tooling-kibi-dedupe). Each case builds a small graph in
# memory and calls kb_dedupe_graph_triples/1, the step kb_attach runs before indexing: a graph holding an entity and
# a relationship twice keeps one copy of each triple and reports the counts, on stderr and in kb_dedupe_report/2,
# which the kibi-cli patch prints for the next kibi command; a clean graph is left alone and reports nothing.
# Needs node_modules (task kibi:install).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

core="$root/node_modules/kibi-core/src/kb.pl"
resolver="$root/node_modules/kibi-cli/dist/prolog/swipl-resolver.js"
if [ ! -f "$core" ] || [ ! -f "$resolver" ]; then
  echo "kibi-dedupe.test.sh: kibi is not installed; run task kibi:install" >&2
  exit 1
fi
# npm ci applies the patch; an install from before the patch existed lacks it.
if ! grep -q 'kb_dedupe_graph_triples' "$core" || [ ! -f "$root/patches/kibi-core+0.31.0.patch" ]; then
  echo "kibi-dedupe.test.sh: the installed kibi-core lacks patches/kibi-core+0.31.0.patch; run task kibi:install" >&2
  exit 1
fi
# kibi's resolver names the SWI-Prolog it runs (its bundled build) and that build's home.
node --input-type=module -e "
  const { resolveSwipl } = await import(process.argv[1]);
  const r = resolveSwipl();
  console.log(r.bin);
  console.log(r.home ?? '');
" "$resolver" >"$work/swipl"
swipl="$(sed -n 1p "$work/swipl")"
home="$(sed -n 2p "$work/swipl")"

# The Prolog side: DUPLICATES=yes asserts one entity's two triples and one relationship triple twice, next to a
# triple stored once. It prints the triple count before and after, the distinct count, and the report.
cat >"$work/case.pl" <<'EOF'
:- initialization(main, main).
main :-
    getenv('KB_PL', KbPl), use_module(KbPl), use_module(library(semweb/rdf_db)),
    G = test_graph,
    rdf_global_id(kb:type, Type),
    Triples = [ t('urn-kibi:REQ-a', Type, literal(req)),
                t('urn-kibi:REQ-a', 'urn-kibi:title', literal('A')),
                t('urn-kibi:REQ-a', 'urn-kibi:specified_by', 'urn-kibi:SCEN-a') ],
    forall(member(t(S, P, O), Triples), rdf_assert(S, P, O, G:1)),
    rdf_assert('urn-kibi:SCEN-a', Type, literal(scenario), G:1),
    % rdf_db keeps an identical triple twice only under another source line; that is how this case stores one twice.
    (   getenv('DUPLICATES', yes)
    ->  forall(member(t(S, P, O), Triples), rdf_assert(S, P, O, G:2))
    ;   true
    ),
    aggregate_all(count, rdf(_, _, _, G), Before),
    kb:kb_dedupe_graph_triples(G),
    aggregate_all(count, rdf(_, _, _, G), After),
    findall(S-P-O, rdf(S, P, O, G), All), sort(All, Distinct), length(Distinct, DistinctCount),
    (   kb:kb_dedupe_report(R, E) -> Report = R/E ; Report = none ),
    format("before=~d after=~d distinct=~d report=~w~n", [Before, After, DistinctCount, Report]).
EOF

# run builds the graph (with duplicates or without) and keeps stdout and stderr apart.
run() {
  env SWI_HOME_DIR="$home" KB_PL="$core" DUPLICATES="$1" "$swipl" -q "$work/case.pl" >"$work/out" 2>"$work/err"
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  out | /' "$work/out" >&2
    sed 's/^/  err | /' "$work/err" >&2
    failures=$((failures + 1))
  fi
}

run yes && rc=0 || rc=$?
check "a graph with an entity and a relationship stored twice keeps one copy of each triple" \
  "$([ "$rc" -eq 0 ] && grep -qx 'before=7 after=4 distinct=4 report=3/1' "$work/out" && echo yes || echo no)"
check "the removal is printed on stderr with both counts" \
  "$(grep -qF 'kibi (local patch): removed 3 duplicate triples (1 entities stored twice) from the published store' "$work/err" && echo yes || echo no)"

run no && rc=0 || rc=$?
check "a clean graph is left alone and reports nothing" \
  "$([ "$rc" -eq 0 ] && grep -qx 'before=4 after=4 distinct=4 report=none' "$work/out" && ! grep -q 'local patch' "$work/err" && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "kibi-dedupe.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "kibi-dedupe.test.sh: all $cases checks passed"

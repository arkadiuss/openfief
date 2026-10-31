#!/usr/bin/env bash
# Ingests the kibi.proof-run.v1 artifacts that `kibi prove` wrote in CI's end-to-end jobs, one per cluster, into
# this checkout's knowledge base (task kibi:ingest; docs/dev/pipeline.md#ci-on-pull-requests-and-main). The docs
# job runs it on the same commit before it builds the site, so the requirements report shows the receipts of the
# real end-to-end runs. kibi derives every receipt itself: `kibi ingest-proof` checks that the artifact was made
# from this workspace snapshot with the configured command, and appends the receipts. A test's receipts must stay
# in the order their runs finished, so the artifacts are ingested oldest first. Anything kibi refuses fails the
# script.
# Usage: kibi-ingest-proofs.sh <kibi executable> <directory holding the artifacts, at any depth>
set -euo pipefail

kibi="$1"
dir="$2"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

find "$dir" -type f -name '*.json' | sort >"$work/runs"
if [ ! -s "$work/runs" ]; then
  echo "kibi-ingest-proofs: no proof-run artifacts under $dir" >&2
  exit 1
fi

# Oldest first: each artifact's run.finished_at, then its path for a stable order.
while IFS= read -r run; do
  printf '%s %s\n' "$(gojq -r .run.finished_at "$run")" "$run"
done <"$work/runs" | sort | cut -d' ' -f2- >"$work/ordered"

while IFS= read -r run; do
  echo "kibi-ingest-proofs: $run: integration $(gojq -r .integration "$run"), run $(gojq -r .run.outcome "$run"), finished $(gojq -r .run.finished_at "$run")"
  gojq '{snapshot: .code_snapshot, artifact: .}' "$run" >"$work/input.json"
  if ! "$kibi" ingest-proof --input "$work/input.json" </dev/null >"$work/out.json"; then
    echo "kibi-ingest-proofs: kibi refused $run:" >&2
    cat "$work/out.json" >&2
    exit 1
  fi
  if ! gojq -e '.status == "success"' "$work/out.json" >/dev/null; then
    echo "kibi-ingest-proofs: kibi refused $run:" >&2
    cat "$work/out.json" >&2
    exit 1
  fi
  gojq -r '.data.results[]? | "kibi-ingest-proofs:   \(.testId): \(.outcome)"' "$work/out.json"
done <"$work/ordered"

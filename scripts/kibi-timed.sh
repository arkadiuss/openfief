#!/usr/bin/env bash
# Runs one kibi command and says how long it took (docs:gen's coverage query; docs/dev/kibi.md#setup-notes). kibi's
# engine stops any single query at 120 seconds, and the docs build asks for the whole coverage report in one query,
# so a knowledge base that slowly grows, or a kibi defect that makes a query slower, would fail the build without
# warning one day. This script prints the duration on stderr every time and warns once it reaches half that limit
# (60 s): under GitHub Actions as a workflow warning, which shows on the run's summary page. Everything this script
# prints goes to stderr (the runner reads workflow commands there too), so the command's stdout, which docs:gen
# writes to a file, stays exactly kibi's; its stderr and exit status pass through unchanged.
# Usage: kibi-timed.sh <kibi executable> <kibi arguments...>
set -uo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: kibi-timed.sh <kibi executable> <kibi arguments...>" >&2
  exit 2
fi
kibi="$1"
shift
limit=120
warn_at=$((limit / 2))

start="$(date +%s)"
"$kibi" "$@"
status=$?
took=$(($(date +%s) - start))

echo "kibi $1 took ${took} s (kibi stops a query at ${limit} s)" >&2
if [ "$took" -ge "$warn_at" ]; then
  message="kibi $1 took ${took} s, half or more of kibi's ${limit} s query limit: find what made it slower before it fails the build (docs/dev/kibi.md)"
  if [ "${GITHUB_ACTIONS:-}" = true ]; then
    echo "::warning title=kibi query time::${message}" >&2
  else
    echo "warning: ${message}" >&2
  fi
fi
exit "$status"

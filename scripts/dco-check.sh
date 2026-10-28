#!/usr/bin/env bash
# DCO check (D-5): every non-merge commit in BASE..HEAD carries "Signed-off-by: <author name> <author email>".
# Usage: dco-check.sh <base> <head>
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: dco-check.sh <base> <head>" >&2
  exit 2
fi

# unsigned prints each non-merge commit in the range whose sign-offs don't include its author.
unsigned() {
  local commit author
  for commit in $(git rev-list --no-merges "$1..$2"); do
    author="$(git log -1 --format='%an <%ae>' "$commit")"
    if ! git log -1 --format='%(trailers:key=Signed-off-by,valueonly)' "$commit" | grep -qxF "$author"; then
      git log -1 --format='%h %s (author: %an <%ae>)' "$commit"
    fi
  done
}

missing="$(unsigned "$1" "$2")"
if [ -n "$missing" ]; then
  echo "dco: these commits lack a sign-off by their author (git commit -s, see D-5):" >&2
  echo "$missing" >&2
  exit 1
fi
echo "dco: every commit is signed off"

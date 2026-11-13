#!/usr/bin/env bash
# DCO check (D-5): every non-merge commit in BASE..HEAD carries "Signed-off-by: <author name> <author email>".
# With --message, it checks the message of the commit being made instead, for the commit-msg hook: the message must
# carry the sign-off of the commit's author (git var GIT_AUTHOR_IDENT), unless a merge is being committed.
# Usage: dco-check.sh <base> <head> | dco-check.sh --message <message-file>
set -euo pipefail

if [ "$#" -eq 2 ] && [ "$1" = "--message" ]; then
  if [ -e "$(git rev-parse --git-path MERGE_HEAD)" ]; then
    exit 0
  fi
  author="$(git var GIT_AUTHOR_IDENT | sed -E 's/ [0-9]+ [-+][0-9]{4}$//')"
  if ! git interpret-trailers --parse <"$2" | sed -n 's/^Signed-off-by: //p' | grep -qxF "$author"; then
    echo "dco: this commit's message lacks a sign-off by its author, $author (git commit -s, see D-5)" >&2
    exit 1
  fi
  exit 0
fi

if [ "$#" -ne 2 ]; then
  echo "usage: dco-check.sh <base> <head> | dco-check.sh --message <message-file>" >&2
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

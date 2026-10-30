#!/usr/bin/env bash
# Tests for scripts/ci-images-select.sh in a throwaway git repository with a fake cosign and a fake go: which
# earlier build it finds for each image and which diff it hands tools/affected; the tool's decisions are its Go
# tests (kibi: TEST-tooling-ci-images-select). Needs gojq on PATH (task test:script adds .bin).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
unset GIT_DIR GIT_INDEX_FILE GIT_WORK_TREE GIT_PREFIX GIT_OBJECT_DIRECTORY
failures=0
cases=0
d1='sha256:1111111111111111111111111111111111111111111111111111111111111111'
d2='sha256:2222222222222222222222222222222222222222222222222222222222222222'

mkdir -p "$work/bin"
# The fake go answers go list with an empty listing and, for go run, copies the input directory to $work/seen
# and prints fixed outputs.
cat >"$work/bin/go" <<'EOF'
#!/usr/bin/env bash
case "$1" in
  list) echo '{}' ;;
  run)
    echo "go $*" >>"$FAKE_LOG"
    while [ "$#" -gt 0 ]; do
      if [ "$1" = -in ]; then rm -rf "$FAKE_SEEN" && cp -r "$2" "$FAKE_SEEN"; fi
      shift
    done
    echo 'images=[]'
    echo 'carry=[]'
    ;;
  *) exit 2 ;;
esac
EOF
# The fake cosign "verifies" only the refs in FAKE_SIGNED (ref=digest), and only for the identity of
# images.yml on main of example/openfief.
cat >"$work/bin/cosign" <<'EOF'
#!/usr/bin/env bash
echo "cosign $2" >>"$FAKE_LOG"
case "$*" in *"--certificate-identity https://github.com/example/openfief/.github/workflows/images.yml@refs/heads/main"*) ;; *) exit 1 ;; esac
for entry in $FAKE_SIGNED; do
  if [ "${entry%%=*}" = "$2" ]; then
    printf '[{"critical":{"image":{"docker-manifest-digest":"%s"}}}]\n' "${entry#*=}"
    exit 0
  fi
done
exit 1
EOF
chmod +x "$work/bin/go" "$work/bin/cosign"

repo="$work/repo"
git init -q "$repo"
git -C "$repo" config user.email test@example.com
git -C "$repo" config user.name Test
mkdir -p "$repo/cmd/core" "$repo/plugins/ui"
touch "$repo/cmd/core/Dockerfile" "$repo/plugins/ui/Dockerfile"
commit() {
  echo "$1" >"$repo/$2"
  git -C "$repo" add . && git -C "$repo" commit -qm "$1"
  git -C "$repo" rev-parse HEAD
}
c1="$(commit one docs.md)"
c2="$(commit two cmd/core/main.go)"
c3="$(commit three docs.md)"

# select runs ci-images-select.sh in the repository; extra environment comes first in "$@".
select_images() {
  : >"$work/log"
  rm -rf "$work/seen"
  local mode="$1" arg="$2"
  shift 2
  (cd "$repo" && env PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_SEEN="$work/seen" "$@" \
    bash "$root/scripts/ci-images-select.sh" ghcr.io/Example/openfief "$mode" "$arg" .) >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" "$work/log" >&2
    [ ! -f "$work/seen/builds" ] || sed 's/^/  builds| /' "$work/seen/builds" >&2
    failures=$((failures + 1))
  fi
}

# same prints yes when the file holds exactly the expected text.
same() {
  if [ "$(cat "$1" 2>/dev/null)" = "$2" ]; then echo yes; else echo no; fi
}

img=ghcr.io/example/openfief
select_images reuse example/openfief FAKE_SIGNED="$img/core:sha-${c1:0:7}=$d1 $img/core:sha-${c2:0:7}=$d2" && rc=0 || rc=$?
check "reuse runs and prints the tool's outputs" "$([ "$rc" -eq 0 ] && grep -qx 'carry=\[\]' "$work/out" && echo yes || echo no)"
check "reuse finds the newest signed build of each image, and none for an image never signed" \
  "$(same "$work/seen/builds" "core $c2 $d2")"
check "reuse stops at the first verified commit" "$(if grep -q "core:sha-${c1:0:7}" "$work/log"; then echo no; else echo yes; fi)"
check "reuse searches HEAD first" "$(if head -1 "$work/log" | grep -q "core:sha-${c3:0:7}"; then echo yes; else echo no; fi)"
check "reuse hands the tool the changes since that build" "$(if tr '\0' '\n' <"$work/seen/diff/$c2" | grep -qx docs.md; then echo yes; else echo no; fi)"
check "reuse runs the tool in reuse mode" "$(yes=no; grep -q -- '-mode reuse' "$work/log" && yes=yes; echo $yes)"

select_images reuse example/openfief FAKE_SIGNED="$img/core:sha-${c3:0:7}=$d1" && rc=0 || rc=$?
check "a rerun finds this commit's own build" "$(same "$work/seen/builds" "core $c3 $d1")"

select_images base "$c1" && rc=0 || rc=$?
check "base lists every image against the base commit" "$(same "$work/seen/builds" "$(printf 'core %s -\nui %s -' "$c1" "$c1")")"
check "base hands the tool the changes since the base" \
  "$(if tr '\0' '\n' <"$work/seen/diff/$c1" | grep -qx cmd/core/main.go; then echo yes; else echo no; fi)"
check "base verifies nothing" "$(if grep -q cosign "$work/log"; then echo no; else echo yes; fi)"

select_images all - && rc=0 || rc=$?
check "all lists no earlier build" "$([ "$rc" -eq 0 ] && [ ! -s "$work/seen/builds" ] && echo yes || echo no)"

for bad in "reuse not-a-repo" "base nope" "nope -"; do
  select_images $bad && rc=0 || rc=$?
  check "refused: $bad" "$([ "$rc" -ne 0 ] && echo yes || echo no)"
done

if [ "$failures" -gt 0 ]; then
  echo "ci-images-select.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "ci-images-select.test.sh: all $cases checks passed"

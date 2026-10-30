#!/usr/bin/env bash
# Tests for scripts/image-carry.sh with a fake crane and cosign over a fake registry in a temporary directory; no
# registry or Sigstore call (kibi: TEST-tooling-image-carry). Needs gojq on PATH (task test:script adds .bin).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
image='ghcr.io/example/openfief/core'
old='sha256:1111111111111111111111111111111111111111111111111111111111111111'
other='sha256:2222222222222222222222222222222222222222222222222222222222222222'
commit='abcdef0123456789abcdef0123456789abcdef01'

# The fake registry is a directory of files named after <image>:<tag>, each holding a digest. The fake crane
# writes a tag; FAKE_CRANE_DIGEST makes it write another digest, like a registry that moved the tag meanwhile.
mkdir -p "$work/bin"
cat >"$work/bin/crane" <<'EOF'
#!/usr/bin/env bash
echo "crane $*" >>"$FAKE_LOG"
[ "$1" = tag ] || exit 2
src="$2" tag="$3"
printf '%s\n' "${FAKE_CRANE_DIGEST:-${src#*@}}" >"$FAKE_REG/$(printf '%s' "${src%@*}:$tag" | tr '/:' '__')"
EOF
# The fake cosign resolves a tag through the fake registry and "verifies" only the digests in FAKE_SIGNED,
# and only for the identity of images.yml on main of example/openfief.
cat >"$work/bin/cosign" <<'EOF'
#!/usr/bin/env bash
echo "cosign $*" >>"$FAKE_LOG"
[ "$1" = verify ] || exit 2
ref="$2"
case "$*" in *"--certificate-identity https://github.com/example/openfief/.github/workflows/images.yml@refs/heads/main"*) ;; *) exit 1 ;; esac
if [[ "$ref" == *@* ]]; then
  digest="${ref#*@}"
else
  file="$FAKE_REG/$(printf '%s' "$ref" | tr '/:' '__')"
  [ -f "$file" ] || { echo "no such tag" >&2; exit 1; }
  digest="$(cat "$file")"
fi
for signed in $FAKE_SIGNED; do
  if [ "$signed" = "$digest" ]; then
    printf '[{"critical":{"image":{"docker-manifest-digest":"%s"}}}]\n' "$digest"
    exit 0
  fi
done
echo "no matching signatures" >&2
exit 1
EOF
chmod +x "$work/bin/crane" "$work/bin/cosign"

# carry runs image-carry.sh on a fresh fake registry; extra environment comes first in "$@".
carry() {
  rm -rf "$work/reg" && mkdir -p "$work/reg"
  : >"$work/log"
  env PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_REG="$work/reg" "$@" \
    bash "$root/scripts/image-carry.sh" "$image" "$old" "$commit" example/openfief >"$work/out" 2>&1
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

# yes_if prints yes when the command succeeds.
yes_if() {
  if "$@" >/dev/null 2>&1; then echo yes; else echo no; fi
}

tagged() { cat "$work/reg/$(printf '%s' "$image:$1" | tr '/:' '__')" 2>/dev/null; }

carry FAKE_SIGNED="$old" && rc=0 || rc=$?
check "a signed digest is carried" "$([ "$rc" -eq 0 ] && echo yes || echo no)"
check "sha-<commit> points at the carried digest" "$([ "$(tagged sha-abcdef0)" = "$old" ] && echo yes || echo no)"
check "main points at the carried digest" "$([ "$(tagged main)" = "$old" ] && echo yes || echo no)"
check "the tag is verified after tagging" "$(yes_if grep -q "cosign verify $image:sha-abcdef0 " "$work/log")"

carry FAKE_SIGNED="$other" && rc=0 || rc=$?
check "an unsigned digest is refused" "$([ "$rc" -ne 0 ] && echo yes || echo no)"
check "an unsigned digest is never tagged" "$(yes_if bash -c "! grep -q '^crane' '$work/log'")"

carry FAKE_SIGNED="$old $other" FAKE_CRANE_DIGEST="$other" && rc=0 || rc=$?
check "a tag that verifies to another digest fails" "$([ "$rc" -ne 0 ] && grep -q "not the carried" "$work/out" && echo yes || echo no)"

rm -rf "$work/reg" && mkdir -p "$work/reg" && : >"$work/log"
for args in "$image notadigest $commit example/openfief" "$image $old xyz example/openfief" "$image $old $commit not-a-repo" "UPPER/case $old $commit example/openfief"; do
  if env PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_REG="$work/reg" FAKE_SIGNED="$old" \
    bash "$root/scripts/image-carry.sh" $args >"$work/out" 2>&1; then rc=0; else rc=$?; fi
  check "malformed arguments are refused: $args" "$([ "$rc" -ne 0 ] && grep -q usage "$work/out" && echo yes || echo no)"
done
check "malformed arguments call nothing" "$(yes_if bash -c "[ ! -s '$work/log' ]")"

if [ "$failures" -gt 0 ]; then
  echo "image-carry.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "image-carry.test.sh: all $cases checks passed"

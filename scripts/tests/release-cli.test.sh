#!/usr/bin/env bash
# Tests for scripts/release-cli.sh with stand-ins for go and cyclonedx-gomod (kibi: TEST-tooling-release-cli): it
# must build the openfief CLI for exactly Linux and macOS on amd64 and arm64, statically, stamped with the version
# `openfief version` prints and the commit (-buildvcs), write the SBOM, and list every file in SHA256SUMS, and refuse
# a version that isn't one. Offline and fast: no real build.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/repo"
failures=0

# go records each build's platform, CGO setting, -buildvcs and -ldflags and writes a fake binary to -o.
cat >"$work/bin/go" <<'EOF'
#!/usr/bin/env bash
out="" ldflags="" vcs=""
while [ $# -gt 0 ]; do
  case "$1" in
    -o) out="$2" ;;
    -ldflags=*) ldflags="${1#-ldflags=}" ;;
    -buildvcs=*) vcs="${1#-buildvcs=}" ;;
  esac
  shift
done
echo "$GOOS/$GOARCH cgo=$CGO_ENABLED vcs=$vcs ldflags=$ldflags" >>"$CALLS"
echo "binary for $GOOS/$GOARCH" >"$out"
EOF
# cyclonedx-gomod writes a fake SBOM to -output.
cat >"$work/bin/cyclonedx-gomod" <<'EOF'
#!/usr/bin/env bash
while [ $# -gt 0 ]; do
  [ "$1" = -output ] && echo '{"bomFormat":"CycloneDX"}' >"$2"
  shift
done
EOF
chmod +x "$work/bin/go" "$work/bin/cyclonedx-gomod"

cd "$work/repo"
if ! out="$(CALLS="$work/calls" PATH="$work/bin:$PATH" bash "$root/scripts/release-cli.sh" 2026.11.0 dist/cli 2>&1)"; then
  echo "FAIL: release-cli: $out" >&2
  failures=$((failures + 1))
fi
stamp='vcs=true ldflags=-s -w -X main.version=2026.11.0'
want="linux/amd64 cgo=0 $stamp
linux/arm64 cgo=0 $stamp
darwin/amd64 cgo=0 $stamp
darwin/arm64 cgo=0 $stamp"
if [ "$(cat "$work/calls")" != "$want" ]; then
  echo "FAIL: builds: want $want, got $(cat "$work/calls")" >&2
  failures=$((failures + 1))
fi
want="openfief-2026.11.0-darwin-amd64
openfief-2026.11.0-darwin-arm64
openfief-2026.11.0-linux-amd64
openfief-2026.11.0-linux-arm64
openfief-2026.11.0.cdx.json"
got="$(awk '{print $2}' dist/cli/openfief-2026.11.0-SHA256SUMS)"
if [ "$got" != "$want" ]; then
  echo "FAIL: SHA256SUMS lists: want $want, got $got" >&2
  failures=$((failures + 1))
fi
if ! (cd dist/cli && sha256sum -c --quiet openfief-2026.11.0-SHA256SUMS); then
  echo "FAIL: SHA256SUMS does not verify" >&2
  failures=$((failures + 1))
fi

# A version reaches -ldflags, so anything but a version's characters is refused before any build.
: >"$work/calls"
if CALLS="$work/calls" PATH="$work/bin:$PATH" bash "$root/scripts/release-cli.sh" '2026.11.0 -X main.x=y' dist/bad >/dev/null 2>&1 ||
  [ -s "$work/calls" ]; then
  echo "FAIL: a version with a space was built" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-cli: 4 checks passed"

#!/usr/bin/env bash
# Tests for scripts/release-cli.sh with stand-ins for go and cyclonedx-gomod (kibi: TEST-tooling-release-cli): it
# must build the openfief CLI for exactly Linux and macOS on amd64 and arm64, statically, write the SBOM, and list
# every file in SHA256SUMS. Offline and fast: no real build.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/repo"
failures=0

# go records each build's platform and CGO setting and writes a fake binary to -o.
cat >"$work/bin/go" <<'EOF'
#!/usr/bin/env bash
out=""
while [ $# -gt 0 ]; do
  [ "$1" = -o ] && out="$2"
  shift
done
echo "$GOOS/$GOARCH cgo=$CGO_ENABLED" >>"$CALLS"
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
want="linux/amd64 cgo=0
linux/arm64 cgo=0
darwin/amd64 cgo=0
darwin/arm64 cgo=0"
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

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-cli: 3 checks passed"

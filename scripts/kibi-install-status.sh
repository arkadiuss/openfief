#!/usr/bin/env bash
# Says whether node_modules still matches what task kibi:install would install (docs/dev/kibi.md#setup-notes): the
# status check of that task, and the stamp it writes after a successful install.
#
# Task's own sources/generates checksum missed a merge that moved kibi-cli and kibi-core to new pins and changed the
# patches, so the old kibi stayed installed and the docs build ran for minutes on it. Instead this compares what is
# installed with what is pinned:
#   - every kibi-* package in package-lock.json (top level) is installed at exactly the pinned version, and
#   - node_modules/.kibi-install-stamp, written by `stamp` after `npm ci` and its patch-package postinstall, lists the
#     sha256 of each file in patches/ as it was applied, and the versions then installed, and both still match. A package
#     reinstalled alone (npm drops its patch) changes no pin but fails the second part; a patch edited, added or
#     removed after the install fails the first.
# Usage: kibi-install-status.sh check|stamp [project root, default .]
# `check` exits 0 when up to date; otherwise it says why on stderr and exits 1, so a task's `status:` runs the install.
# `stamp` writes the stamp. Needs node, as the install itself does.
set -euo pipefail

mode="${1:-}"
root="${2:-.}"
if [ "$mode" != check ] && [ "$mode" != stamp ]; then
  echo "usage: kibi-install-status.sh check|stamp [project root]" >&2
  exit 2
fi
stamp="$root/node_modules/.kibi-install-stamp"

# pinned prints "<package> <version>" for each kibi package package-lock.json pins, sorted. An optional package (kibi's
# SWI-Prolog binaries for the other platforms) counts only where it is installed: npm skips it on this machine.
pinned() {
  node -e '
    const fs = require("fs");
    const lock = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
    for (const [key, pkg] of Object.entries(lock.packages || {})) {
      const m = /^node_modules\/(kibi-[^/]+)$/.exec(key);
      if (!m) continue;
      if (pkg.optional && !fs.existsSync(process.argv[2] + "/node_modules/" + m[1] + "/package.json")) continue;
      console.log(m[1], pkg.version);
    }
  ' "$root/package-lock.json" "$root" | LC_ALL=C sort
}

# installed prints "<package> <installed version>" for the packages pinned() names, "missing" when one isn't installed.
installed() {
  local name version
  while read -r name _; do
    version=missing
    if [ -f "$root/node_modules/$name/package.json" ]; then
      version="$(node -e 'console.log(JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).version)' \
        "$root/node_modules/$name/package.json")"
    fi
    echo "$name $version"
  done < <(pinned)
}

# patches prints "<sha256> <file name>" for each patch in patches/, sorted by name.
patches() {
  local f
  for f in "$root"/patches/*.patch; do
    [ -e "$f" ] || continue
    echo "$(sha256sum "$f" | cut -d' ' -f1) $(basename "$f")"
  done
}

# state is the stamp's content: the applied patches, then the installed versions.
state() {
  patches | sed 's/^/patch /'
  installed | sed 's/^/installed /'
}

if [ "$mode" = stamp ]; then
  # A stamp for an install that doesn't match the pins would hide the mismatch, so refuse to write one.
  if ! diff -q <(pinned) <(installed) >/dev/null; then
    echo "kibi-install-status: not stamping, the installed kibi packages differ from the pins" >&2
    exit 1
  fi
  state >"$stamp"
  exit 0
fi

if ! diff_out="$(diff <(pinned) <(installed))"; then
  echo "kibi-install-status: installed kibi packages differ from package-lock.json (< pinned, > installed):" >&2
  echo "$diff_out" >&2
  exit 1
fi
if [ ! -f "$stamp" ]; then
  echo "kibi-install-status: no install stamp, so the patches are not known to be applied" >&2
  exit 1
fi
if ! diff_out="$(diff "$stamp" <(state))"; then
  echo "kibi-install-status: patches or installed versions changed since the last install (< stamped, > now):" >&2
  echo "$diff_out" >&2
  exit 1
fi

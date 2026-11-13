---
title: kibi-install-status.sh compares installed kibi versions and patches with the pins
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:41:59.260Z'
sourceFile: scripts/tests/kibi-install-status.test.sh
verification_scope: integration
id: TEST-tooling-kibi-install-tracks-pins
type: test
---
Runs scripts/kibi-install-status.sh on a temporary project tree with a fake package-lock.json, node_modules and patches, no network and no npm: a stamped install at the pinned versions is up to date; an older kibi-cli, a lock that moved on, a missing package, or an installed optional package at the wrong version fails naming both versions; an optional package npm skipped and a package that is not kibi's are ignored; an edited, added or removed patch and a missing stamp fail; stamp refuses an install that differs from the pins; an unknown mode is a usage error.

---
title: The release bundle's core runs SpiceDB, and a bundle whose core doesn't is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-22T08:41:13.073Z'
id: SCEN-tooling-release-bundle-authz-spicedb
type: scenario
---
Given the real deploy/, when scripts/release-bundle.sh renders a release's bundle, then the install file's core runs --authz=spicedb and no other --authz, reads SpiceDB's endpoint and plaintext setting from its environment and mounts the Secret openfief-spicedb-key as an optional volume; given a deploy/ whose core runs --authz=policy, or no --authz at all, the script fails naming what the core runs and writes no file. Assumes the core's container is the one named core, as deploy/core names it.
---
title: A published release lists its plugins in the catalog its own binaries embed
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:58.404Z'
id: SCEN-tooling-release-catalog-entries
type: scenario
---
Given a release the author started by hand with dry_run false, when the plugin images are pushed and signed, then each plugins/<name> enters the catalog with its signed digest, the release version and OpenFief's CI as publisher, on a release commit from which the operator, core and CLI are built and which the tag names; a dry run adds nothing.
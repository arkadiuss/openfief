---
title: openfief version prints the stamped version, or dev, and the build's commit
status: active
sourceFile: cmd/openfief/version_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T16:08:49.653Z'
id: TEST-cmd-openfief-version
type: test
---
A table test of versionLine over synthetic build settings: a release version with a clean revision, dev, a modified tree, no vcs settings and an empty revision. TestVersionCommand runs the command's entry point in the test binary, which is unstamped, so it must print one line starting openfief dev commit, and refuse an argument with the usage error.

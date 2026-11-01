---
title: release-version.sh accepts calendar versions per channel and refuses the rest
status: active
sourceFile: scripts/tests/release-version.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:46.976Z'
id: TEST-tooling-release-version
type: test
verification_scope: integration
---
Runs scripts/release-version.sh in a throwaway repository over a table of channel and version pairs: valid stable and beta versions pass; a v prefix, leading zeros, a month out of range, the wrong channel suffix, an unknown channel and an existing tag fail.

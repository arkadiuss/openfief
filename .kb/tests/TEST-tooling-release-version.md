---
title: release-version.sh accepts calendar versions per channel, alpha included, and refuses the rest
status: active
sourceFile: scripts/tests/release-version.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:46.976Z'
id: TEST-tooling-release-version
type: test
verification_scope: integration
---
Runs scripts/release-version.sh in a throwaway repository over a table of 27 channel and version pairs: valid stable, beta and alpha versions pass; a v prefix, leading zeros, a month out of range, beta.0 and alpha.0, an alpha without K or in upper case, the wrong channel suffix, an unknown channel and an existing stable or alpha tag fail.
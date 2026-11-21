---
title: release-fetch.sh fetches and verifies a release as the guide says, and fails on each broken check
status: active
sourceFile: scripts/tests/release-fetch.test.sh
verification_scope: integration
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-tooling-release-fetch
type: test
---
Runs scripts/release-fetch.sh against a fake release with stand-ins for gh and cosign: a release whose signatures verify is fetched into an empty directory with its CLI executable and its demo-seed image recorded by the verified digest; a wrong identity accepted, a changed CLI or SBOM, a missing asset, an unsigned image, an image by tag, a third party's image without its digest, an unsigned demo-seed image, a CloudNativePG manifest that is not the pinned one and a directory that is not empty each fail it. A false pass would be a fake cosign that accepts everything, which the wrong-identity case rules out.
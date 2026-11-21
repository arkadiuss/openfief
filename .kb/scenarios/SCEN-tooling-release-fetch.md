---
title: release-fetch.sh fetches a release only when every signature, checksum and pin holds
status: active
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-tooling-release-fetch
type: scenario
---
Given a release's assets and images that its workflows signed, when release-fetch.sh runs for that version into an empty directory, then the bundle, the CLI and its checksums are in the directory, the CLI is executable, and demo-seed.image names the demo-seed image by the digest cosign verified; and when a signature names a wrong identity, a file's checksum or signature fails, an image of ours is unsigned or not pinned by digest, the CloudNativePG manifest is not the pinned one, or the demo-seed image is unsigned, the script fails with what failed.

Assumptions: cosign, gh and the registry are stand-ins in the script test, so no registry or Sigstore is called; the real release is checked by the upgrade run of the end-to-end suite.

---
title: The published release installs as the guide says, or the smoke test fails
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T16:42:33.468Z'
id: SCEN-tooling-release-smoke
type: scenario
---
Given a published release, when release-smoke.yml runs for its version, then it downloads the assets with gh, verifies the CLI, checksums, SBOM and every image with the guide's cosign identities, sees cosign refuse a wrong identity, checks that openfief version names the release and its tagged commit, checks the CloudNativePG checksum, applies CloudNativePG, the CRDs, the database and the install, waits for the rollouts, installs the UI with the downloaded CLI and creates the first passkey through openfief enroll. A wrong identity accepted, a bad signature or checksum, an image by tag, another CLI version, no enrolment link or a failed passkey fails it.

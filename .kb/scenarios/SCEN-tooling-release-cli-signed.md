---
title: Every release ships the signed CLI
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:02:00.575Z'
id: SCEN-tooling-release-cli-signed
type: scenario
---
Given any release, beta or stable, when the cli job runs, then it builds static openfief binaries for linux and darwin on amd64 and arm64 from the release commit with a CycloneDX SBOM and SHA256SUMS; on a real run each file gets a keyless cosign bundle and all of them are attached to the GitHub release, a pre-release for a beta.
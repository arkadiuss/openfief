---
title: The script release-fetch.sh must download a published release's bundle and CLI from its GitHub release into an empty directory, verify the CLI, its SBOM and its checksums as the release workflow on main signed them and every image the install runs as the images workflow on main signed it, check that cosign refuses a wrong identity, check CloudNativePG's manifest against its pinned sha256 and record the release's demo-seed image by the digest cosign verified, failing at the first check that does not hold.
status: open
firmness: default
required_evidence: unverified
text_ref: docs/dev/sdlc.md#releases
tags:
  - release
  - e2e
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
semantic_text: The script release-fetch.sh must download a published release's bundle and CLI from its GitHub release into an empty directory, verify the CLI, its SBOM and its checksums as the release workflow on main signed them and every image the install runs as the images workflow on main signed it, check that cosign refuses a wrong identity, check CloudNativePG's manifest against its pinned sha256 and record the release's demo-seed image by the digest cosign verified, failing at the first check that does not hold.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b5b5652c4f82b67c26eaae743625df9544fec590523a704174cdb0551481dd9e
logic_claims:
  - CLAIM-4FAE3D4AC1C3BB40
semantic_inventory:
  - claim_key: CLAIM-4FAE3D4AC1C3BB40
    claim_text: The script release-fetch.sh must download a published release's bundle and CLI from its GitHub release into an empty directory, verify the CLI, its SBOM and its checksums as the release workflow on main signed them and every image the install runs as the images workflow on main signed it, check that cosign refuses a wrong identity, check CloudNativePG's manifest against its pinned sha256 and record the release's demo-seed image by the digest cosign verified, failing at the first check that does not hold
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 508
    subjects:
      - tooling.release_fetch
id: REQ-tooling-release-fetch
type: req
---
The script release-fetch.sh must download a published release's bundle and CLI from its GitHub release into an empty directory, verify the CLI, its SBOM and its checksums as the release workflow on main signed them and every image the install runs as the images workflow on main signed it, check that cosign refuses a wrong identity, check CloudNativePG's manifest against its pinned sha256 and record the release's demo-seed image by the digest cosign verified, failing at the first check that does not hold.

## Context

The design session asked for an end-to-end test that upgrades the published 2026.11.0-alpha.1 in place, and that test must install the real released artifact verified the way the install guide and the post-release smoke test verify it. The smoke test's download and verification steps are part of a script that also installs, so this script does the fetching alone and reuses release-verify.sh for the signatures.

## Source

> install a published release the way the guide tells an owner to, verifying its signature the same way

Source: the design session's brief for exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21

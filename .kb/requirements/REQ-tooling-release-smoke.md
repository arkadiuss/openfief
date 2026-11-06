---
title: After every published release, CI must install that release from its GitHub release assets on a fresh kind cluster exactly as the install guide says, through the CLI install of the UI and the first passkey, and fail when any step does not do what the guide says.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: After every published release, CI must install that release from its GitHub release assets on a fresh kind cluster exactly as the install guide says, through the CLI install of the UI and the first passkey, and fail when any step does not do what the guide says.
logic_claims:
  - CLAIM-F62483E029A26DC6
semantic_clauses:
  - After every published release, CI must install that release from its GitHub release assets on a fresh kind cluster exactly as the install guide says, through the CLI install of the UI and the first passkey, and fail when any step does not do what the guide says.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 58fe44666371fbc43d208ccd2b4674a30de7d10136f7f3b1410c41659f487680
semantic_inventory:
  - claim_key: CLAIM-F62483E029A26DC6
    claim_text: After every published release, CI must install that release from its GitHub release assets on a fresh kind cluster exactly as the install guide says, through the CLI install of the UI and the first passkey, and fail when any step does not do what the guide says
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-11-06T16:42:36.390Z'
id: REQ-tooling-release-smoke
type: req
---
After every published release, CI must install that release from its GitHub release assets on a fresh kind cluster exactly as the install guide says, through the CLI install of the UI and the first passkey, and fail when any step does not do what the guide says.

## Context

After 2026.11.0-alpha.1 the orchestrator checked the install guide's steps 1 to 4 by hand but could not run steps 5 to 7 without a cluster, so it asked for a workflow that runs every step against a published release, as an owner would. The end-to-end suite installs the bundle of the checkout and the CLI built from it (docs/dev/testing.md), so only a run after publishing covers the release's own assets, their signatures and the downloaded CLI. release.yml calls release-smoke.yml after publish, and it also starts by hand for any version. The negative identity checks prove cosign really pins the identity the guide gives (D-181).

## Source

> Every release, an alpha included, ships signed images and the bootstrap install with the CRDs, every image in it pinned to the digest that was signed. An end-to-end test installs OpenFief from that bundle.

docs/dev/sdlc.md#releases

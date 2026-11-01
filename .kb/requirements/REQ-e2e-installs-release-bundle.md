---
title: An end-to-end test must install OpenFief from the release bundle rendered from the checkout's images, with its files applied unchanged, and the operator, the egress proxy and the core installed from it must become available and the core must answer the read API.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: An end-to-end test must install OpenFief from the release bundle rendered from the checkout's images, with its files applied unchanged, and the operator, the egress proxy and the core installed from it must become available and the core must answer the read API.
logic_claims:
  - CLAIM-AED974648EFA1E95
semantic_clauses:
  - An end-to-end test must install OpenFief from the release bundle rendered from the checkout's images, with its files applied unchanged, and the operator, the egress proxy and the core installed from it must become available and the core must answer the read API.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8ac54e28b3060d6f058324e1443eea598a1b1ec8991b42bbf4ec4b4992469f9a
semantic_inventory:
  - claim_key: CLAIM-AED974648EFA1E95
    claim_text: An end-to-end test must install OpenFief from the release bundle rendered from the checkout's images, with its files applied unchanged, and the operator, the egress proxy and the core installed from it must become available and the core must answer the read API
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-11-01T13:51:20.597Z'
id: REQ-e2e-installs-release-bundle
type: req
---
An end-to-end test must install OpenFief from the release bundle rendered from the checkout's images, with its files applied unchanged, and the operator, the egress proxy and the core installed from it must become available and the core must answer the read API.

## Context

The author's release rule in docs/dev/sdlc.md, added with the first release as an alpha, so a release can't ship an install that doesn't start. Before it, the end-to-end suite installed only the demo's manifests. Modeled by Claude Code for release readiness item R1 (impl/release-bundle); the author has not reviewed it.

## Source

> Every release, an alpha included, ships signed images and the bootstrap install with the CRDs, every image in it pinned to the digest that was signed. An end-to-end test installs OpenFief from that bundle.

docs/dev/sdlc.md#releases

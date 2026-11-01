---
title: Every release must ship the CRDs and the bootstrap install of the operator, the egress proxy and the core as a bundle in which every image is pinned to the digest that release signed, rendered deterministically, and must refuse to render it when an image has no digest or its signature doesn't verify.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: Every release must ship the CRDs and the bootstrap install of the operator, the egress proxy and the core as a bundle in which every image is pinned to the digest that release signed, rendered deterministically, and must refuse to render it when an image has no digest or its signature doesn't verify.
logic_claims:
  - CLAIM-D25001BAE9F7BD57
semantic_clauses:
  - Every release must ship the CRDs and the bootstrap install of the operator, the egress proxy and the core as a bundle in which every image is pinned to the digest that release signed, rendered deterministically, and must refuse to render it when an image has no digest or its signature doesn't verify.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3bdbb2bdb3e2a161c7c5a4113ca9383a08d1952d05bd79d2c9a711eb774df442
semantic_inventory:
  - claim_key: CLAIM-D25001BAE9F7BD57
    claim_text: Every release must ship the CRDs and the bootstrap install of the operator, the egress proxy and the core as a bundle in which every image is pinned to the digest that release signed, rendered deterministically, and must refuse to render it when an image has no digest or its signature doesn't verify
    role: normative
    status: modeled
    span:
      start: 0
      end: 300
origin:
  kind: agent
  recorded_at: '2026-11-01T13:51:00.724Z'
id: REQ-tooling-release-bundle-pinned
type: req
---
Every release must ship the CRDs and the bootstrap install of the operator, the egress proxy and the core as a bundle in which every image is pinned to the digest that release signed, rendered deterministically, and must refuse to render it when an image has no digest or its signature doesn't verify.

## Context

The author's release rule in docs/dev/sdlc.md, added with the first release as an alpha: until now the stable tarball shipped deploy/ with bases that name images without a tag or digest. Plugins aren't in the bundle: the catalog pins their signed digests (D-88) and the CLI installs them (D-46). D-17 pins images by digest. Modeled by Claude Code for release readiness item R1 (impl/release-bundle); the author has not reviewed it.

## Source

> Every release, an alpha included, ships signed images and the bootstrap install with the CRDs, every image in it pinned to the digest that was signed. An end-to-end test installs OpenFief from that bundle.

docs/dev/sdlc.md#releases

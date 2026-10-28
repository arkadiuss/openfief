---
title: The egress proxy must cache an accepted token review for at most a TTL below the projected token lifetime and never past the token's expiry, never cache a refused token, keep the cache bounded, and hold only token digests.
status: open
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
semantic_text: The egress proxy must cache an accepted token review for at most a TTL below the projected token lifetime and never past the token's expiry, never cache a refused token, keep the cache bounded, and hold only token digests.
logic_claims:
  - CLAIM-56B24DA2A3A94846
semantic_clauses:
  - The egress proxy must cache an accepted token review for at most a TTL below the projected token lifetime and never past the token's expiry, never cache a refused token, keep the cache bounded, and hold only token digests.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a468b94e47f5501f8230013288ac21a379f1b2f2b698f0ed26e83bacaa786a95
semantic_inventory:
  - claim_key: CLAIM-56B24DA2A3A94846
    claim_text: The egress proxy must cache an accepted token review for at most a TTL below the projected token lifetime and never past the token's expiry, never cache a refused token, keep the cache bounded, and hold only token digests
    role: normative
    status: modeled
    span:
      start: 0
      end: 221
origin:
  kind: agent
  recorded_at: '2026-10-28T08:47:07.265Z'
id: REQ-plugins-egress-token-review-cache
type: req
---
The egress proxy must cache an accepted token review for at most a TTL below the projected token lifetime and never past the token's expiry, never cache a refused token, keep the cache bounded, and hold only token digests.

## Context

D-63 says the role function checks the token with TokenReview and caches the result, so the API server is not asked on every CONNECT. The orchestrator asked for a bounded TTL below the token expiry; five minutes is used against 3600 s tokens. Trade-off: a token whose pod or ServiceAccount was deleted stays accepted for up to the TTL. The exp claim is read only after TokenReview accepted the token. Modeled by Claude Code for the M0 egress proxy v2 task (D-63 conformance); the author has not reviewed it yet.

## Source

> The egress proxy embeds Smokescreen through `BuildProxy` inside OpenFief's own HTTP server; `StartWithConfig` would exit the process on errors and take over signals. Smokescreen allows every destination when no allow-list is set, so the proxy never runs without one and refuses all destinations for a plugin that has none. Compiled allow-lists are swapped atomically. A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result.

docs/DESIGN.md D-63

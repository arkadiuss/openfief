---
title: A changed egress allow-list must replace the previous one atomically, so each CONNECT is decided against one complete list and the next CONNECT after a change sees the new list.
status: open
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
semantic_text: A changed egress allow-list must replace the previous one atomically, so each CONNECT is decided against one complete list and the next CONNECT after a change sees the new list.
logic_claims:
  - CLAIM-33F1F60CA967B7EA
semantic_clauses:
  - A changed egress allow-list must replace the previous one atomically, so each CONNECT is decided against one complete list and the next CONNECT after a change sees the new list.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 061c607cede0c5dc528fb53bfe101e50ca42c38ebba4342015c1c2fee3353b22
semantic_inventory:
  - claim_key: CLAIM-33F1F60CA967B7EA
    claim_text: A changed egress allow-list must replace the previous one atomically, so each CONNECT is decided against one complete list and the next CONNECT after a change sees the new list
    role: normative
    status: modeled
    span:
      start: 0
      end: 176
origin:
  kind: agent
  recorded_at: '2026-10-28T08:47:18.813Z'
id: REQ-plugins-egress-allow-lists-swapped-atomically
type: req
---
A changed egress allow-list must replace the previous one atomically, so each CONNECT is decided against one complete list and the next CONNECT after a change sees the new list.

## Context

D-63 says compiled allow-lists are swapped atomically: the operator rewrites the ConfigMap as grants change, and a half-applied list could briefly allow a revoked host or refuse a granted one. Refines REQ-AUTO-48C62B395F846700. Modeled by Claude Code for the M0 egress proxy v2 task (D-63 conformance); the author has not reviewed it yet.

## Source

> The egress proxy embeds Smokescreen through `BuildProxy` inside OpenFief's own HTTP server; `StartWithConfig` would exit the process on errors and take over signals. Smokescreen allows every destination when no allow-list is set, so the proxy never runs without one and refuses all destinations for a plugin that has none. Compiled allow-lists are swapped atomically. A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result.

docs/DESIGN.md D-63

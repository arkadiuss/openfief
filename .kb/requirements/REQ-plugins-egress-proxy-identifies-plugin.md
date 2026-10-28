---
title: The egress proxy must identify a plugin only from a Proxy-Authorization Bearer projected token that the TokenReview API authenticates for audience openfief-core as a plugin ServiceAccount, refuse every other request, and never log the token.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The egress proxy must identify a plugin only from a Proxy-Authorization Bearer projected token that the TokenReview API authenticates for audience openfief-core as a plugin ServiceAccount, refuse every other request, and never log the token.
logic_claims:
  - CLAIM-C749997F9CF3C2F0
semantic_clauses:
  - The egress proxy must identify a plugin only from a Proxy-Authorization Bearer projected token that the TokenReview API authenticates for audience openfief-core as a plugin ServiceAccount, refuse every other request, and never log the token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 60bdf6f8271518c23f3cc6ef3eeba9e48ec955eb43d39ecb1bba72b0d1442c32
semantic_inventory:
  - claim_key: CLAIM-C749997F9CF3C2F0
    claim_text: The egress proxy must identify a plugin only from a Proxy-Authorization Bearer projected token that the TokenReview API authenticates for audience openfief-core as a plugin ServiceAccount, refuse every other request, and never log the token
    role: normative
    status: modeled
    span:
      start: 0
      end: 240
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:15.853Z'
id: REQ-plugins-egress-proxy-identifies-plugin
type: req
---
The egress proxy must identify a plugin only from a Proxy-Authorization Bearer projected token that the TokenReview API authenticates for audience openfief-core as a plugin ServiceAccount, refuse every other request, and never log the token.

## Context

D-32 has plugins identify themselves to the proxy with their projected ServiceAccount token through a custom role function, and D-27 fixes the audience openfief-core. The token travels in Proxy-Authorization because the CONNECT request is the only thing the proxy sees; plugins and the SDKs must send it on every CONNECT and re-read the rotating token file. TokenReview checks signature, expiry, binding and audience; the username maps to a plugin only as system:serviceaccount:openfief:plugin-<name>. Refines REQ-AUTO-15828B9727AFA488 and REQ-AUTO-C6544ADAF8F5AE65. Modeled by Claude Code for the M0 egress proxy task (internal/plugins/app, smokescreen and kube adapters, cmd/egress-proxy); the author has not reviewed it yet.

## Source

> Plugins identify themselves with their projected ServiceAccount token, through a custom role function.

docs/DESIGN.md D-32

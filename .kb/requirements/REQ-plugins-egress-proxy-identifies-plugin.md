---
title: The egress proxy must identify a plugin only from a projected token sent as the password of a Basic Proxy-Authorization header, which the TokenReview API authenticates for audience openfief-egress as a plugin ServiceAccount, refuse every other request, and never log the token.
status: open
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
semantic_text: The egress proxy must identify a plugin only from a projected token sent as the password of a Basic Proxy-Authorization header, which the TokenReview API authenticates for audience openfief-egress as a plugin ServiceAccount, refuse every other request, and never log the token.
logic_claims:
  - CLAIM-8A8941C5C74DCF70
semantic_clauses:
  - The egress proxy must identify a plugin only from a projected token sent as the password of a Basic Proxy-Authorization header, which the TokenReview API authenticates for audience openfief-egress as a plugin ServiceAccount, refuse every other request, and never log the token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: de88b2b4ded7d2378f144b91e46269c5aecc702433aad1bb4a68bf9d62d6343c
semantic_inventory:
  - claim_key: CLAIM-8A8941C5C74DCF70
    claim_text: The egress proxy must identify a plugin only from a projected token sent as the password of a Basic Proxy-Authorization header, which the TokenReview API authenticates for audience openfief-egress as a plugin ServiceAccount, refuse every other request, and never log the token
    role: normative
    status: modeled
    span:
      start: 0
      end: 276
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:15.853Z'
id: REQ-plugins-egress-proxy-identifies-plugin
type: req
---
The egress proxy must identify a plugin only from a projected token sent as the password of a Basic Proxy-Authorization header, which the TokenReview API authenticates for audience openfief-egress as a plugin ServiceAccount, refuse every other request, and never log the token.

## Context

D-63 fixes how plugin identity (D-32) reaches the proxy: the plugin's SDK puts its projected token in the proxy URL as the password, which HTTP clients send as Basic Proxy-Authorization on CONNECT, and the role function checks it with TokenReview for audience openfief-egress, separate from the core's openfief-core (D-27) so the proxy never holds a token the core accepts. The user name is ignored. The username maps to a plugin only as system:serviceaccount:openfief:plugin-<name>. Refines REQ-AUTO-15828B9727AFA488. Modeled by Claude Code for the M0 egress proxy v2 task (D-63 conformance); the author has not reviewed it yet.

## Source

> The egress proxy embeds Smokescreen through `BuildProxy` inside OpenFief's own HTTP server; `StartWithConfig` would exit the process on errors and take over signals. Smokescreen allows every destination when no allow-list is set, so the proxy never runs without one and refuses all destinations for a plugin that has none. Compiled allow-lists are swapped atomically. A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result.

docs/DESIGN.md D-63

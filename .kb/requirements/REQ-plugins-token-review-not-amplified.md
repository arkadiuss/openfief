---
title: The core's browser port must accept only sessions and script tokens, and the core and the egress proxy must send TokenReview only tokens shaped like a JSON Web Token and remember a refusal for a short while, so no caller can make them call the API server at will.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The core's browser port must accept only sessions and script tokens, and the core and the egress proxy must send TokenReview only tokens shaped like a JSON Web Token and remember a refusal for a short while, so no caller can make them call the API server at will.
logic_claims:
  - CLAIM-6ACBB95747A6911E
semantic_clauses:
  - The core's browser port must accept only sessions and script tokens, and the core and the egress proxy must send TokenReview only tokens shaped like a JSON Web Token and remember a refusal for a short while, so no caller can make them call the API server at will.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f492883d288707da31986d0254cf140eadcd8fd32b000f2eb1b7bacc549245ca
semantic_inventory:
  - claim_key: CLAIM-6ACBB95747A6911E
    claim_text: The core's browser port must accept only sessions and script tokens, and the core and the egress proxy must send TokenReview only tokens shaped like a JSON Web Token and remember a refusal for a short while, so no caller can make them call the API server at will
    role: normative
    status: modeled
    span:
      start: 0
      end: 262
origin:
  kind: agent
  recorded_at: '2026-11-03T00:11:13.749Z'
id: REQ-plugins-token-review-not-amplified
type: req
---
The core's browser port must accept only sessions and script tokens, and the core and the egress proxy must send TokenReview only tokens shaped like a JSON Web Token and remember a refusal for a short while, so no caller can make them call the API server at will.

## Context

Every token the read API didn't know went to a TokenReview, from either port, and refusals were never cached, so anyone reaching the browser port could turn each request into a call to the API server. D-184 bounds what unauthenticated requests may cost; this fix applies it to TokenReview. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag. No firm requirement for this part of D-184 existed in kibi on this branch when it was built.

## Source

> The database, the egress proxy and the operator each have a NetworkPolicy admitting only their callers. Pages the core proxies from a view plugin carry security headers the core sets: a content security policy, framing denied, no content sniffing and no referrer.

docs/DESIGN.md#D-184

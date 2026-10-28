---
title: Caches Token Reviews Bounded = true
status: active
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: caches_token_reviews_bounded
canonical_key: mcp-kibi-model-requirement:plugins.egress_proxy:caches_token_reviews_bounded:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-56B24DA2A3A94846
claim_text: The egress proxy must cache an accepted token review for at most a TTL below the projected token lifetime and never past the token's expiry, never cache a refused token, keep the cache bounded, and hold only token digests.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:47:02.808Z'
id: FACT-PROP-9A14FBAF3BD7D6E8
type: fact
---

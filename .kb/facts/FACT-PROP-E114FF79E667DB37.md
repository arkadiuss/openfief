---
title: Pins Every Image To Signed Digest = true
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: release.bundle
property_key: pins_every_image_to_signed_digest
canonical_key: mcp-kibi-model-requirement:release.bundle:pins_every_image_to_signed_digest:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D25001BAE9F7BD57
claim_text: Every release must ship the CRDs and the bootstrap install of the operator, the egress proxy and the core as a bundle in which every image is pinned to the digest that release signed, rendered deterministically, and must refuse to render it when an image has no digest or its signature doesn't verify.
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:55.031Z'
id: FACT-PROP-E114FF79E667DB37
type: fact
---

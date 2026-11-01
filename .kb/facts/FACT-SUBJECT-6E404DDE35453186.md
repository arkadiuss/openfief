---
title: Release Bundle
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: release.bundle
canonical_key: release.bundle
claim_key: CLAIM-D25001BAE9F7BD57
claim_text: Every release must ship the CRDs and the bootstrap install of the operator, the egress proxy and the core as a bundle in which every image is pinned to the digest that release signed, rendered deterministically, and must refuse to render it when an image has no digest or its signature doesn't verify.
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:50.172Z'
id: FACT-SUBJECT-6E404DDE35453186
type: fact
---

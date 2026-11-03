---
title: Identity Token Review
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: identity.token_review
canonical_key: identity.token_review
claim_key: CLAIM-6ACBB95747A6911E
claim_text: The core's browser port must accept only sessions and script tokens, and the core and the egress proxy must send TokenReview only tokens shaped like a JSON Web Token and remember a refusal for a short while, so no caller can make them call the API server at will.
origin:
  kind: agent
  recorded_at: '2026-11-03T00:11:08.426Z'
id: FACT-SUBJECT-5F672FBFA74DFF0D
type: fact
---

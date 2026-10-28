---
title: Views Transaction Listing
status: active
text_ref: docs/DESIGN.md#D-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-6
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: views.transaction_listing
canonical_key: views.transaction_listing
claim_key: CLAIM-6F18288550BE7608
claim_text: Transaction listings must be paged in date then source then external ID order with an opaque cursor and at most 500 transactions per page, so a client reads every current transaction exactly once.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:33.365Z'
id: FACT-SUBJECT-393ADA027839BE5B
type: fact
---

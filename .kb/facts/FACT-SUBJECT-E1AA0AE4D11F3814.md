---
title: Ingest Opening Cash
status: active
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.opening_cash
canonical_key: ingest.opening_cash
claim_key: CLAIM-EF69FA5FF51D9578
claim_text: The cash balance the owner enters with the opening upload of a statement that holds none must be booked once as the account's opening cash on its go-live date, and refused with any other upload, for an account already seeded, in another currency than the account's, or by a connector whose statement holds the cash.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:47.542Z'
id: FACT-SUBJECT-E1AA0AE4D11F3814
type: fact
---

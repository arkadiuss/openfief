---
title: Ledger Opening Lot
status: active
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.opening_lot
canonical_key: ledger.opening_lot
claim_key: CLAIM-DB616724C3841645
claim_text: An opening lot must be recorded on the go-live date with the date the lot was first acquired and its cost basis in the original currency, and an open date after go-live must be rejected.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:55.100Z'
id: FACT-SUBJECT-7674F14A65896C4F
type: fact
---
An opening lot: recorded on the go-live date with the date the lot was first acquired and its cost basis in the original currency, an open date after go-live being rejected. Lots added one at a time are all dated on the go-live date when the account is seeded, on one date when not, and refused for an account with a current transaction.
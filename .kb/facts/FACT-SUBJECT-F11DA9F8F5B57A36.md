---
title: Ledger Opening Lots
status: archived
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.opening_lots
canonical_key: ledger.opening_lots
claim_key: CLAIM-18B50858DCCF5E92
claim_text: Opening lots added one at a time to an account must all be dated on its go-live date when it is seeded, on one date when it is not, and be refused for an account not seeded that has a current transaction on or before that date; opening cash must never be added that way.
origin:
  kind: agent
  recorded_at: '2026-10-30T05:01:58.438Z'
id: FACT-SUBJECT-F11DA9F8F5B57A36
type: fact
part_of:
  - ledger.opening_lot
---
The dating of opening lots added one at a time to an account: all on its go-live date when seeded, on one date when not, and refused when the account has a current transaction. It is a part of ledger.opening_lot.
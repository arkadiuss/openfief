---
title: Ledger Lot
status: active
text_ref: docs/DESIGN.md#31-system-of-record-and-views
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-31-system-of-record-and-views
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.lot
canonical_key: ledger.lot
claim_key: CLAIM-E9334EFBE9C77039
claim_text: Disposals must consume an account's lots of an instrument oldest first, by open date and then by the acquiring transaction's key, with an opening lot ordered by its broker-reported open date.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:16.715Z'
id: FACT-SUBJECT-2DE79D898756870D
type: fact
---

Definition not stated: the subject ledger.lot is named by requirements REQ-decision-d34-1-consumed-cost-rounding, REQ-decision-d34-2-exact-remaining-cost, REQ-decision-d35-2-acquisitions-before-disposals, REQ-ledger-fifo, REQ-ledger-lot-cost-split, REQ-ledger-oversell, which link to the subject fact FACT-SUBJECT-2DE79D898756870D through constrains. The people who own that source can say what it is in the product and what it is not.

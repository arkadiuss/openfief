---
title: Ledger Cash Checkpoint Service
status: active
text_ref: docs/DESIGN.md#D-158
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-158
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.cash_checkpoint_service
canonical_key: ledger.cash_checkpoint_service
claim_key: CLAIM-72EE8F3A3344A2BE
claim_text: The CashCheckpointService must admit only the owner's view, refuse an account the owner may not change, and refuse a malformed date, currency or amount, an unknown account or checkpoint, and a checkpoint with nothing to book, storing nothing then.
origin:
  kind: agent
  recorded_at: '2026-11-02T03:50:13.859Z'
id: FACT-SUBJECT-D76CF65945DB4D93
type: fact
---

Definition not stated: the subject ledger.cash_checkpoint_service is named by requirement REQ-ledger-cash-checkpoint-service, which links to the subject fact FACT-SUBJECT-D76CF65945DB4D93 through constrains. The people who own that source can say what it is in the product and what it is not.

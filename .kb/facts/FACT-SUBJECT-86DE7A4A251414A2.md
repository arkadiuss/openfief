---
title: Ledger Rekey
status: active
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.rekey
canonical_key: ledger.rekey
claim_key: CLAIM-5730629E84BFDF73
claim_text: The ledger must re-key an instrument by appending, in one database transaction, a new version under the same key of each current transaction that posts to it, with only its holding postings moved, running the import hooks, carrying its display names over, and changing nothing when run again.
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:19.492Z'
id: FACT-SUBJECT-86DE7A4A251414A2
type: fact
---

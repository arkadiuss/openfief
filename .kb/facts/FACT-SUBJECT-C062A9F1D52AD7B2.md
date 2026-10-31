---
title: Ledger Deletion
status: active
text_ref: docs/DESIGN.md#D-100
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-100
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.deletion
canonical_key: ledger.deletion
claim_key: CLAIM-C465E8F1865B2A99
claim_text: A deletion must be a new version of its transaction's key with no postings that supersedes the current version, keeps every earlier version, leaves the transaction out of current transactions and everything derived from them, and must be refused for a key that has no version.
origin:
  kind: agent
  recorded_at: '2026-10-30T05:01:47.398Z'
id: FACT-SUBJECT-C062A9F1D52AD7B2
type: fact
---

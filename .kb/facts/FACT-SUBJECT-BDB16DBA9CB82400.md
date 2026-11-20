---
title: Ingest Bond Lot Entry
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.bond_lot_entry
canonical_key: ingest.bond_lot_entry
claim_key: CLAIM-6AAC65AA02FC492F
claim_text: The core must take a bond lot only from the owner's owner-scoped session, hand it to the one connector instance bound to its account as a new transactions batch, and import what that instance delivers into that account only, recording the delivered series terms.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:23.033Z'
id: FACT-SUBJECT-BDB16DBA9CB82400
type: fact
---

Definition not stated: the subject ingest.bond_lot_entry is named by requirements REQ-ingest-bond-account-live-on-first-lot, REQ-ingest-bond-lot-delete-deletes-coupons, REQ-ingest-bond-lot-delete-entry, REQ-ingest-bond-lot-delete-excludes-sync, REQ-ingest-bond-lot-delete-supersedes, REQ-ingest-bond-lot-go-live, REQ-ingest-bond-lot-handed-with-account, REQ-ingest-bond-lot-handoff, REQ-ingest-bond-lot-no-go-live, REQ-ingest-bond-terms-entered-with-lot, REQ-ingest-bond-terms-latest-entry-wins, REQ-ingest-bond-terms-owner-overrides, REQ-treasurybonds-lot-delete-deletes-booked, which link to the subject fact FACT-SUBJECT-BDB16DBA9CB82400 through constrains. The people who own that source can say what it is in the product and what it is not.

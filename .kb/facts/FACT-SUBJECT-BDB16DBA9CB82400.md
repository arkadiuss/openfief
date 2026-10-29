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

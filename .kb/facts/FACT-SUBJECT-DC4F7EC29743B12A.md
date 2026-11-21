---
title: Ingest Bond Lot
status: active
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.bond_lot
canonical_key: ingest.bond_lot
claim_key: CLAIM-21FAF5872FD12268
claim_text: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, and can be valued on its purchase date by the Polish bond rules as a cash purchase.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:08.895Z'
id: FACT-SUBJECT-DC4F7EC29743B12A
type: fact
---

Definition not stated: the subject ingest.bond_lot is named by requirements REQ-ingest-bond-lot-exchange-price-by-letter, REQ-ingest-bond-lot-purchase-kinds, REQ-ingest-bond-lot-rules, REQ-ingest-bond-lot-terms-optional, which link to the subject fact FACT-SUBJECT-DC4F7EC29743B12A through constrains. The people who own that source can say what it is in the product and what it is not.

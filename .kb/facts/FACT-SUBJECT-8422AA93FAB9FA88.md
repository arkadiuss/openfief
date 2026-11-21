---
title: Treasurybonds Lot Mapping
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
subject_key: treasurybonds.lot_mapping
canonical_key: treasurybonds.lot_mapping
claim_key: CLAIM-076844F381AA59B7
claim_text: The treasury bond connector must map each valid lot to a deposit of its cost and a buy of its bonds on its purchase date in its configured account, keyed by series and purchase date, and deliver each series' terms once under an instrument whose symbol is the series code.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:15.681Z'
id: FACT-SUBJECT-8422AA93FAB9FA88
type: fact
---

Definition not stated: the subject treasurybonds.lot_mapping is named by requirements REQ-treasurybonds-lot-booked-on-purchase-date, REQ-treasurybonds-lot-idempotent-series-date, REQ-treasurybonds-lot-mapping, REQ-treasurybonds-lot-on-go-live-day-opening, REQ-treasurybonds-lot-on-purchase-date, REQ-treasurybonds-lots-any-order, REQ-treasurybonds-post-go-live-deposit-buy, REQ-treasurybonds-pre-go-live-opening-lot, which link to the subject fact FACT-SUBJECT-8422AA93FAB9FA88 through constrains. The people who own that source can say what it is in the product and what it is not.

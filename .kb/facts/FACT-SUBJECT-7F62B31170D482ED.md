---
title: Treasurybonds Coupons
status: active
text_ref: docs/DESIGN.md#D-117
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-117
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: treasurybonds.coupons
canonical_key: treasurybonds.coupons
claim_key: CLAIM-FADD8F3E5EC210B7
claim_text: In a transactions sync, the treasury-bonds connector must book each coupon that fell due after the account's go-live and on or before today on an open lot the request names as interest, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper.
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:33.242Z'
id: FACT-SUBJECT-7F62B31170D482ED
type: fact
---

Definition not stated: the subject treasurybonds.coupons is named by requirements REQ-treasurybonds-coupons-as-interest, REQ-treasurybonds-coupons-booked-every-sync, REQ-treasurybonds-cpi-years-kept, REQ-treasurybonds-payouts-from-purchase, which link to the subject fact FACT-SUBJECT-7F62B31170D482ED through constrains. The people who own that source can say what it is in the product and what it is not.

---
title: Ledger Payout Setting
status: active
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.payout_setting
canonical_key: ledger.payout_setting
claim_key: CLAIM-C5BEE4A7399CD380
claim_text: Each account must keep a payout setting, cash by default or paid out, with the day it took effect, and what falls due must be paid out only when the setting is paid out and took effect on or before that day.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:54.971Z'
id: FACT-SUBJECT-FD62B4328BE42BDF
type: fact
---

Definition not stated: the subject ledger.payout_setting is named by requirements REQ-ledger-first-payout-setting-whole-history, REQ-ledger-payout-change-from-owner-date, REQ-ledger-payout-setting, REQ-ledger-payout-setting-history, which link to the subject fact FACT-SUBJECT-FD62B4328BE42BDF through constrains. The people who own that source can say what it is in the product and what it is not.

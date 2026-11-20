---
title: Ledger Transfer
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-105-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
text_ref: design-decisions:D-105#1
fact_kind: subject
subject_key: ledger.transfer
canonical_key: ledger.transfer
claim_key: CLAIM-B3882A0F98494B51
claim_text: A withdrawal must be suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first.
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:12.304Z'
id: FACT-SUBJECT-2ED66112542A7725
type: fact
---

Definition not stated: the subject ledger.transfer is named by requirements REQ-ledger-transfer-broken-attention-item, REQ-ledger-transfer-broken-stops-counting, REQ-ledger-transfer-correction, REQ-ledger-transfer-leg-once, REQ-ledger-transfer-legs, REQ-ledger-transfer-link-append-only, REQ-ledger-transfer-suggestion, REQ-ui-transfer-confirmation, which link to the subject fact FACT-SUBJECT-2ED66112542A7725 through constrains. The people who own that source can say what it is in the product and what it is not.

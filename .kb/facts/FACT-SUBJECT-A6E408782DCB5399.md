---
title: Ledger Owner Cash Balance
status: active
text_ref: design-decisions:D-158#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-158-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ledger.owner_cash_balance
canonical_key: ledger.owner_cash_balance
claim_key: CLAIM-5706E766B6474CE1
claim_text: The owner must be able to enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's.
origin:
  kind: agent
  recorded_at: '2026-11-01T12:31:09.364Z'
id: FACT-SUBJECT-A6E408782DCB5399
type: fact
---

Definition not stated: the subject ledger.owner_cash_balance is named by requirements REQ-ledger-derived-cash-never-overwritten, REQ-ledger-owner-cash-difference-booked, REQ-ledger-owner-cash-difference-default-deposit, REQ-ledger-owner-cash-mismatch-attention-item, REQ-ledger-owner-cash-stored-as-assertion, REQ-ledger-owner-enters-institution-cash, which link to the subject fact FACT-SUBJECT-A6E408782DCB5399 through constrains. The people who own that source can say what it is in the product and what it is not.

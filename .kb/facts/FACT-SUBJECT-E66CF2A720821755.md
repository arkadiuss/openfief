---
title: Ledger Accounts
status: archived
text_ref: design-decisions:D-23#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-23-1
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: ledger.accounts
canonical_key: ledger.accounts
claim_key: CLAIM-05CFC61DD32EF841
claim_text: Accounts must be seeded with opening lots from each broker's current lot-level positions.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:12:38.841Z'
id: FACT-SUBJECT-E66CF2A720821755
type: fact
part_of:
  - ledger.account
---
Seeding accounts with opening lots from each broker's current lot-level positions, atomically and refusing a reseed. It is a part of ledger.account, which covers what an account is and must have.
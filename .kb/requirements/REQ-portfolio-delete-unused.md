---
title: The owner must be able to delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-120-1
  - lane:strict
text_ref: design-decisions:D-120#1
origin:
  kind: import
  ref: design-decisions:D-120#1
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The owner must be able to delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions.
logic_claims:
  - CLAIM-8B016E93DCA36B72
semantic_clauses:
  - The owner must be able to delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2658637cea75f3add80b9891bf6cf7abaf267ba54634d2c28198abbabc9964d1
semantic_inventory:
  - claim_key: CLAIM-8B016E93DCA36B72
    claim_text: The owner must be able to delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions
    role: normative
    status: modeled
    span:
      start: 0
      end: 124
id: REQ-portfolio-delete-unused
type: req
---
The owner must be able to delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions.

## Context

D-71 and D-72 let the owner add and rename sleeves, sleeve kinds and accounts but didn't say whether any can be removed, so the implementation deleted nothing and an unwanted default sleeve could only be renamed or left empty (question #5 on docs/implementation-questions.md). On 2026-10-31 the author allowed deleting what nothing depends on. Nothing the ledger holds is ever deleted (D-18, REQ-AUTO-22E70B6530E994B1), which is why only an account with no transactions can go.

## Source

> The owner can delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions, and remap a broker account whose OpenFief account has no transactions from it. Nothing the ledger holds is ever deleted (D-18).

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-120#1

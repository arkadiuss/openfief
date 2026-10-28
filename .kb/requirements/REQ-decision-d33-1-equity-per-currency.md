---
title: Each account must have an equity ledger account in each currency, besides the D-19 ledger accounts.
status: open
text_ref: design-decisions:D-33#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-33-1
  - lane:strict
semantic_text: Each account must have an equity ledger account in each currency, besides the D-19 ledger accounts.
logic_claims:
  - CLAIM-7BE4B2C75F5DE93C
semantic_clauses:
  - Each account must have an equity ledger account in each currency, besides the D-19 ledger accounts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ebf4877e5c0de39b1883bc64ffbaaa483b6f261a02a12759b77bd53f60b1aba3
semantic_inventory:
  - claim_key: CLAIM-7BE4B2C75F5DE93C
    claim_text: Each account must have an equity ledger account in each currency, besides the D-19 ledger accounts
    role: normative
    status: modeled
    span:
      start: 0
      end: 98
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-33#1
  recorded_at: '2026-10-28T06:27:53.218Z'
id: REQ-decision-d33-1-equity-per-currency
type: req
---
Each account must have an equity ledger account in each currency, besides the D-19 ledger accounts.

## Context

The author decided D-33 in docs/DESIGN.md section 3.3; this requirement imports its claim D-33#1 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Besides D-19's accounts, each account has an equity ledger account per currency: the counterpart for deposits, withdrawals and opening lots.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-33#1

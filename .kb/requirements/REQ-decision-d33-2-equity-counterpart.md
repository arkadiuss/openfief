---
title: An account's equity ledger account must be the counterpart posting for deposits, withdrawals and opening lots.
status: open
text_ref: design-decisions:D-33#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-33-2
  - lane:strict
semantic_text: An account's equity ledger account must be the counterpart posting for deposits, withdrawals and opening lots.
logic_claims:
  - CLAIM-63054D54560E9B98
semantic_clauses:
  - An account's equity ledger account must be the counterpart posting for deposits, withdrawals and opening lots.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3980c16263c7ecd7fef1acd622477925fc471ba5676ef1e0c48d3605b9d10629
semantic_inventory:
  - claim_key: CLAIM-63054D54560E9B98
    claim_text: An account's equity ledger account must be the counterpart posting for deposits, withdrawals and opening lots
    role: normative
    status: modeled
    span:
      start: 0
      end: 109
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-33#2
  recorded_at: '2026-10-28T06:28:07.717Z'
id: REQ-decision-d33-2-equity-counterpart
type: req
---
An account's equity ledger account must be the counterpart posting for deposits, withdrawals and opening lots.

## Context

The author decided D-33 in docs/DESIGN.md section 3.3; this requirement imports its claim D-33#2 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Besides D-19's accounts, each account has an equity ledger account per currency: the counterpart for deposits, withdrawals and opening lots.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-33#2

---
title: The owner must confirm or reject each suggested transfer in the UI.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-105-2
  - lane:strict
text_ref: design-decisions:D-105#2
origin:
  kind: import
  ref: design-decisions:D-105#2
  recorded_at: '2026-10-29T10:10:05.766Z'
semantic_text: The owner must confirm or reject each suggested transfer in the UI.
logic_claims:
  - CLAIM-D93486C282FFA739
semantic_clauses:
  - The owner must confirm or reject each suggested transfer in the UI.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ec1eaa3596b0cb25418f3e6d3e89bbac2f51f8689b69da72cfec6b7413daf010
semantic_inventory:
  - claim_key: CLAIM-D93486C282FFA739
    claim_text: The owner must confirm or reject each suggested transfer in the UI
    role: normative
    status: modeled
    span:
      start: 0
      end: 66
id: REQ-ui-transfer-confirmation
type: req
---
The owner must confirm or reject each suggested transfer in the UI.

## Context

D-105 identifies transfers between the owner's own accounts by suggestion and confirmation: OpenFief suggests a match (REQ-ledger-transfer-suggestion), the owner decides in the UI, and only a confirmed transfer leaves the time-weighted return's cash flows (REQ-portfolio-twr-confirmed-transfers-only). The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-105#2

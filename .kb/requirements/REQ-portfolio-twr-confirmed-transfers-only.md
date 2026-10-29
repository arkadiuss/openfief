---
title: Only confirmed transfers must be left out of the time-weighted return's cash flows.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-105-4
  - lane:strict
text_ref: design-decisions:D-105#4
origin:
  kind: import
  ref: design-decisions:D-105#4
  recorded_at: '2026-10-29T10:11:24.441Z'
semantic_text: Only confirmed transfers must be left out of the time-weighted return's cash flows.
logic_claims:
  - CLAIM-7EA84DB6754B430C
semantic_clauses:
  - Only confirmed transfers must be left out of the time-weighted return's cash flows.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 70bac66e3e6376c2743cf80011e1dfdf8307c76a05ec363234f30df4dca4b386
semantic_inventory:
  - claim_key: CLAIM-7EA84DB6754B430C
    claim_text: Only confirmed transfers must be left out of the time-weighted return's cash flows
    role: normative
    status: modeled
    span:
      start: 0
      end: 82
id: REQ-portfolio-twr-confirmed-transfers-only
type: req
---
Only confirmed transfers must be left out of the time-weighted return's cash flows.

## Context

D-61 measures the portfolio's line as a time-weighted return and says transfers between the owner's own accounts are not cash flows. D-105 makes that depend on the owner: a suggested transfer (REQ-ledger-transfer-suggestion) leaves the return's cash flows only once the owner confirms it (REQ-ui-transfer-confirmation). The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-105#4

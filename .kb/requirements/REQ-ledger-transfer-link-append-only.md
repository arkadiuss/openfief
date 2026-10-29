---
title: A confirmed transfer link must be stored in the ledger, append-only.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-105-3
  - lane:strict
text_ref: design-decisions:D-105#3
origin:
  kind: import
  ref: design-decisions:D-105#3
  recorded_at: '2026-10-29T10:10:45.274Z'
semantic_text: A confirmed transfer link must be stored in the ledger, append-only.
logic_claims:
  - CLAIM-3AD3B22B11016BC4
semantic_clauses:
  - A confirmed transfer link must be stored in the ledger, append-only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c88d49787b972452faa4efa6a1793b80115a466e4c79eef383062183686fc43a
semantic_inventory:
  - claim_key: CLAIM-3AD3B22B11016BC4
    claim_text: A confirmed transfer link must be stored in the ledger, append-only
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
id: REQ-ledger-transfer-link-append-only
type: req
---
A confirmed transfer link must be stored in the ledger, append-only.

## Context

D-105 has the owner confirm or reject each suggested transfer (REQ-ui-transfer-confirmation) and keeps each confirmed link in the ledger, append-only. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-105#3

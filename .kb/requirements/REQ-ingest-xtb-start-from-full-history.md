---
title: An XTB account must be able to start from an export covering its whole history.
status: open
text_ref: design-decisions:D-136#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-136-1
  - lane:strict
semantic_text: An XTB account must be able to start from an export covering its whole history.
logic_claims:
  - CLAIM-1EBB20CA677D42DA
semantic_clauses:
  - An XTB account must be able to start from an export covering its whole history.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 68fa8bddc0f023c6ed8e210786e4c285457a9591db0553b636b2b85215044189
semantic_inventory:
  - claim_key: CLAIM-1EBB20CA677D42DA
    claim_text: An XTB account must be able to start from an export covering its whole history
    role: normative
    status: modeled
    span:
      start: 0
      end: 78
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-136#1
  recorded_at: '2026-11-01T07:39:44.306Z'
id: REQ-ingest-xtb-start-from-full-history
type: req
---
An XTB account must be able to start from an export covering its whole history.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. XTB shut down its public API in 2025, so XTB accounts arrive as xStation export files (D-9). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-136#1

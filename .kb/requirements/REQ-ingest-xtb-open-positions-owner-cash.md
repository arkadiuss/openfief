---
title: An XTB account started from a current export's open positions must take the cash balance the owner enters once.
status: open
text_ref: design-decisions:D-136#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-136-3
  - lane:strict
semantic_text: An XTB account started from a current export's open positions must take the cash balance the owner enters once.
logic_claims:
  - CLAIM-5B7AF2802C93B0D1
semantic_clauses:
  - An XTB account started from a current export's open positions must take the cash balance the owner enters once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 220bdb6c2c25a60e51120d4741d27255e2e6752c388f155e215403e6daa734e7
semantic_inventory:
  - claim_key: CLAIM-5B7AF2802C93B0D1
    claim_text: An XTB account started from a current export's open positions must take the cash balance the owner enters once
    role: normative
    status: modeled
    span:
      start: 0
      end: 110
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-136#3
  recorded_at: '2026-11-01T07:39:57.605Z'
id: REQ-ingest-xtb-open-positions-owner-cash
type: req
---
An XTB account started from a current export's open positions must take the cash balance the owner enters once.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. A current export's open positions don't give the account's cash, so the owner enters it once when the account starts. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-136#3

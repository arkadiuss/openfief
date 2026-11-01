---
title: An XTB account must be able to start from the open positions of a current export, which seed one opening lot per position at its open time and price.
status: open
text_ref: design-decisions:D-136#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-136-2
  - lane:strict
semantic_text: An XTB account must be able to start from the open positions of a current export, which seed one opening lot per position at its open time and price.
logic_claims:
  - CLAIM-367BEA96255AC17A
semantic_clauses:
  - An XTB account must be able to start from the open positions of a current export, which seed one opening lot per position at its open time and price.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e6600247eb405d416c8761df611f26f97c9bef69ee68b7faaa1f05160c6a6c88
semantic_inventory:
  - claim_key: CLAIM-367BEA96255AC17A
    claim_text: An XTB account must be able to start from the open positions of a current export, which seed one opening lot per position at its open time and price
    role: normative
    status: modeled
    span:
      start: 0
      end: 148
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-136#2
  recorded_at: '2026-11-01T07:39:50.747Z'
id: REQ-ingest-xtb-start-from-open-positions
type: req
---
An XTB account must be able to start from the open positions of a current export, which seed one opening lot per position at its open time and price.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. This is D-23's seeding from current positions for an XTB account whose full history the owner doesn't upload. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-136#2

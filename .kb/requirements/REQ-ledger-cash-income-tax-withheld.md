---
title: Cash income must be booked with any tax withheld.
status: open
text_ref: design-decisions:D-152#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-152-2
  - lane:strict
semantic_text: Cash income must be booked with any tax withheld.
logic_claims:
  - CLAIM-4BC03CC4C056F490
semantic_clauses:
  - Cash income must be booked with any tax withheld.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b941fecf52ae06e01fdfd6bb11f18f46b864c3bcf400769a53d905e5c34181ce
semantic_inventory:
  - claim_key: CLAIM-4BC03CC4C056F490
    claim_text: Cash income must be booked with any tax withheld
    role: normative
    status: modeled
    span:
      start: 0
      end: 48
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-152#2
  recorded_at: '2026-11-01T12:29:20.554Z'
id: REQ-ledger-cash-income-tax-withheld
type: req
---
Cash income must be booked with any tax withheld.

## Context

The author recorded it on 2026-11-01. D-19 gives the ledger withholding-tax postings, which carry the tax withheld from income. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The ledger has a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result, booked with any tax withheld. Performance counts it as return, not as money paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-152#2

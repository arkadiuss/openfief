---
title: The cash balance of each Ghostfolio account the plugin owns must follow OpenFief's cash, so Ghostfolio's totals include it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-108-3
  - lane:strict
text_ref: design-decisions:D-108#3
origin:
  kind: import
  ref: design-decisions:D-108#3
  recorded_at: '2026-10-29T10:26:34.402Z'
semantic_text: The cash balance of each Ghostfolio account the plugin owns must follow OpenFief's cash, so Ghostfolio's totals include it.
logic_claims:
  - CLAIM-EBBE0371940CC77C
semantic_clauses:
  - The cash balance of each Ghostfolio account the plugin owns must follow OpenFief's cash, so Ghostfolio's totals include it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 897fa92b2ed1aa65bcd3d3abf6d2f00df9652d838054f8fa12a608f7744fd5f8
semantic_inventory:
  - claim_key: CLAIM-EBBE0371940CC77C
    claim_text: The cash balance of each Ghostfolio account the plugin owns must follow OpenFief's cash, so Ghostfolio's totals include it
    role: normative
    status: modeled
    span:
      start: 0
      end: 122
id: REQ-views-ghostfolio-cash-balance
type: req
---
The cash balance of each Ghostfolio account the plugin owns must follow OpenFief's cash, so Ghostfolio's totals include it.

## Context

D-4 has the Ghostfolio plugin run a reconcile loop over the Ghostfolio accounts it created, and D-64 says how it syncs. D-108 records, as the author's decision, how the plugin as built books dividends, standalone fees, cash and listed instruments in Ghostfolio. The author amended D-64 and recorded D-108 after the design session that answered the implementation's questions.

## Source

> The Ghostfolio plugin books a dividend as quantity 1 at the gross amount, with the withheld tax as its fee, and a standalone fee against a manual "OpenFief fees" asset per currency. Each of its Ghostfolio accounts' cash balance follows OpenFief's cash, so Ghostfolio's totals include it. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-108#3

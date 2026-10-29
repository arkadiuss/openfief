---
title: The Ghostfolio plugin must book a dividend as quantity 1 at the gross amount, with the withheld tax as its fee.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-108-1
  - lane:strict
text_ref: design-decisions:D-108#1
origin:
  kind: import
  ref: design-decisions:D-108#1
  recorded_at: '2026-10-29T10:25:11.822Z'
semantic_text: The Ghostfolio plugin must book a dividend as quantity 1 at the gross amount, with the withheld tax as its fee.
logic_claims:
  - CLAIM-F9E76E84DBB933AA
semantic_clauses:
  - The Ghostfolio plugin must book a dividend as quantity 1 at the gross amount, with the withheld tax as its fee.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 50d342e57c21c10b7d9e1d70d9e60423817fef8a0237196048be8ed885a10ac9
semantic_inventory:
  - claim_key: CLAIM-F9E76E84DBB933AA
    claim_text: The Ghostfolio plugin must book a dividend as quantity 1 at the gross amount, with the withheld tax as its fee
    role: normative
    status: modeled
    span:
      start: 0
      end: 110
id: REQ-views-ghostfolio-dividend-booking
type: req
---
The Ghostfolio plugin must book a dividend as quantity 1 at the gross amount, with the withheld tax as its fee.

## Context

D-4 has the Ghostfolio plugin run a reconcile loop over the Ghostfolio accounts it created, and D-64 says how it syncs. D-108 records, as the author's decision, how the plugin as built books dividends, standalone fees, cash and listed instruments in Ghostfolio. The author amended D-64 and recorded D-108 after the design session that answered the implementation's questions.

## Source

> The Ghostfolio plugin books a dividend as quantity 1 at the gross amount, with the withheld tax as its fee, and a standalone fee against a manual "OpenFief fees" asset per currency. Each of its Ghostfolio accounts' cash balance follows OpenFief's cash, so Ghostfolio's totals include it. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-108#1

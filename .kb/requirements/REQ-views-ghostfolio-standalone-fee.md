---
title: The Ghostfolio plugin must book a standalone fee against a manual "OpenFief fees" asset per currency.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-108-2
  - lane:strict
text_ref: design-decisions:D-108#2
origin:
  kind: import
  ref: design-decisions:D-108#2
  recorded_at: '2026-10-29T10:25:52.237Z'
semantic_text: The Ghostfolio plugin must book a standalone fee against a manual "OpenFief fees" asset per currency.
logic_claims:
  - CLAIM-6369ECFCDC83523F
semantic_clauses:
  - The Ghostfolio plugin must book a standalone fee against a manual "OpenFief fees" asset per currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2a149b8ff880d7b7fdb38d3406af2ce70ae22b7c060cd0acd694e5debcb22dbb
semantic_inventory:
  - claim_key: CLAIM-6369ECFCDC83523F
    claim_text: The Ghostfolio plugin must book a standalone fee against a manual "OpenFief fees" asset per currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
id: REQ-views-ghostfolio-standalone-fee
type: req
---
The Ghostfolio plugin must book a standalone fee against a manual "OpenFief fees" asset per currency.

## Context

D-4 has the Ghostfolio plugin run a reconcile loop over the Ghostfolio accounts it created, and D-64 says how it syncs. D-108 records, as the author's decision, how the plugin as built books dividends, standalone fees, cash and listed instruments in Ghostfolio. The author amended D-64 and recorded D-108 after the design session that answered the implementation's questions.

## Source

> The Ghostfolio plugin books a dividend as quantity 1 at the gross amount, with the withheld tax as its fee, and a standalone fee against a manual "OpenFief fees" asset per currency. Each of its Ghostfolio accounts' cash balance follows OpenFief's cash, so Ghostfolio's totals include it. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-108#2

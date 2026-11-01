---
title: Every day since a hand-entered bond account's first purchase must be valued from the series' terms and the published readings.
status: open
text_ref: design-decisions:D-98#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-6
  - lane:strict
semantic_text: Every day since a hand-entered bond account's first purchase must be valued from the series' terms and the published readings.
logic_claims:
  - CLAIM-74859E5F87C584B5
semantic_clauses:
  - Every day since a hand-entered bond account's first purchase must be valued from the series' terms and the published readings.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 866eecad6b072eeed6077db90789db0820b205da93923cef98093083ade55f49
semantic_inventory:
  - claim_key: CLAIM-74859E5F87C584B5
    claim_text: Every day since a hand-entered bond account's first purchase must be valued from the series' terms and the published readings
    role: normative
    status: modeled
    span:
      start: 0
      end: 125
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-98#6
  recorded_at: '2026-11-01T07:38:08.942Z'
id: REQ-treasurybonds-valued-every-day-since-first-purchase
type: req
---
Every day since a hand-entered bond account's first purchase must be valued from the series' terms and the published readings.

## Context

On 2026-11-01 the author rewrote D-98 and removed D-119: a hand-entered bond lot is booked on its real purchase date, because real purchase dates are the standard and keep the ledger a record of what happened. There are no opening lots at go-live for hand-entered bonds any more. The published readings are the CPI and NBP reference rate that D-94's sources deliver, and the series' terms come from D-97. D-36 stores the valuations. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#6

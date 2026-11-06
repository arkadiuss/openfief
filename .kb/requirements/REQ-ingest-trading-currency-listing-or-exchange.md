---
title: An XTB position's trading currency must be the currency of the confirmed ticker's listing or, when the listing gives none, the currency of the exchange its market suffix names.
status: open
text_ref: design-decisions:D-159#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-159-2
  - lane:strict
semantic_text: An XTB position's trading currency must be the currency of the confirmed ticker's listing or, when the listing gives none, the currency of the exchange its market suffix names.
logic_claims:
  - CLAIM-93E9CC51A173B944
semantic_clauses:
  - An XTB position's trading currency must be the currency of the confirmed ticker's listing or, when the listing gives none, the currency of the exchange its market suffix names.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6d23e75bf39a062b7b19a6eee5a4dd351238c7b103286d84b73aaa63c9536c7e
semantic_inventory:
  - claim_key: CLAIM-93E9CC51A173B944
    claim_text: An XTB position's trading currency must be the currency of the confirmed ticker's listing or, when the listing gives none, the currency of the exchange its market suffix names
    role: normative
    status: modeled
    span:
      start: 0
      end: 175
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-159#2
  recorded_at: '2026-11-02T00:54:17.616Z'
id: REQ-ingest-trading-currency-listing-or-exchange
type: req
---
An XTB position's trading currency must be the currency of the confirmed ticker's listing or, when the listing gives none, the currency of the exchange its market suffix names.

## Why this changed

On 2026-11-01 the author recorded D-159. REQ-ingest-ticker-trading-currency, a default requirement the implementation agent recorded while building D-159 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-159 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. The owner confirms the ticker's ISIN and listing once (D-116); D-170 has the owner pick the currency instead when that exchange quotes in more than one. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB position's open price is in its instrument's trading currency, which the export doesn't name. It is the currency of the confirmed ticker's listing (D-116), or, when the listing gives none, the currency of the exchange its market suffix names, and the opening lot's cost is booked in it. A price quoted in a minor unit, such as London's pence, is converted to the major unit.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-159#2

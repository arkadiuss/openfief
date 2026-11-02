---
title: An XTB position's open price must be taken as in its instrument's trading currency, which the export doesn't name.
status: open
text_ref: design-decisions:D-159#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-159-1
  - lane:strict
semantic_text: An XTB position's open price must be taken as in its instrument's trading currency, which the export doesn't name.
logic_claims:
  - CLAIM-BA3235265F3200A9
semantic_clauses:
  - An XTB position's open price must be taken as in its instrument's trading currency, which the export doesn't name.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 064729ef45595526a5a73d5929243eb053e9ff1466514b6dec267ece41def160
semantic_inventory:
  - claim_key: CLAIM-BA3235265F3200A9
    claim_text: An XTB position's open price must be taken as in its instrument's trading currency, which the export doesn't name
    role: normative
    status: modeled
    span:
      start: 0
      end: 113
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-159#1
  recorded_at: '2026-11-02T00:54:09.473Z'
id: REQ-ingest-xtb-open-price-trading-currency
type: req
---
An XTB position's open price must be taken as in its instrument's trading currency, which the export doesn't name.

## Why this changed

On 2026-11-01 the author recorded D-159. REQ-ingest-xtb-opening-lot-trading-currency, a default requirement the implementation agent recorded while building D-159 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-159 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. The decision gives the fact behind it: an xStation export names no currency for a position's open price. Reason not stated beyond that. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB position's open price is in its instrument's trading currency, which the export doesn't name. It is the currency of the confirmed ticker's listing (D-116), or, when the listing gives none, the currency of the exchange its market suffix names, and the opening lot's cost is booked in it. A price quoted in a minor unit, such as London's pence, is converted to the major unit.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-159#1

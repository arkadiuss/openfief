---
title: A price quoted in a minor unit, such as London's pence, must be converted to the major unit.
status: open
text_ref: design-decisions:D-159#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-159-4
  - lane:strict
semantic_text: A price quoted in a minor unit, such as London's pence, must be converted to the major unit.
logic_claims:
  - CLAIM-EF0D2AFEFACFC278
semantic_clauses:
  - A price quoted in a minor unit, such as London's pence, must be converted to the major unit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e203158c7dc9c9015d91e2e819c102d987c73fdc1a85d46bc03c2e293e182f94
semantic_inventory:
  - claim_key: CLAIM-EF0D2AFEFACFC278
    claim_text: A price quoted in a minor unit, such as London's pence, must be converted to the major unit
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-159#4
  recorded_at: '2026-11-02T00:54:29.857Z'
id: REQ-ingest-minor-unit-price-to-major
type: req
---
A price quoted in a minor unit, such as London's pence, must be converted to the major unit.

## Why this changed

On 2026-11-01 the author recorded D-159. REQ-ingest-ticker-trading-currency, a default requirement the implementation agent recorded while building D-159 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-159 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. Amounts are booked in a currency's major unit (D-20). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB position's open price is in its instrument's trading currency, which the export doesn't name. It is the currency of the confirmed ticker's listing (D-116), or, when the listing gives none, the currency of the exchange its market suffix names, and the opening lot's cost is booked in it. A price quoted in a minor unit, such as London's pence, is converted to the major unit.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-159#4

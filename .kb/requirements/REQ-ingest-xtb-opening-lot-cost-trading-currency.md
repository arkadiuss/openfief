---
title: An XTB opening lot's cost must be booked in its position's trading currency.
status: open
text_ref: design-decisions:D-159#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-159-3
  - lane:strict
semantic_text: An XTB opening lot's cost must be booked in its position's trading currency.
logic_claims:
  - CLAIM-8D7954B098EE79FA
semantic_clauses:
  - An XTB opening lot's cost must be booked in its position's trading currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2d8bb0ea1ffaf65dc8cc5ac8731d681d488979804e0105aac2dda1e3edbbcf15
semantic_inventory:
  - claim_key: CLAIM-8D7954B098EE79FA
    claim_text: An XTB opening lot's cost must be booked in its position's trading currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 75
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-159#3
  recorded_at: '2026-11-02T00:54:23.384Z'
id: REQ-ingest-xtb-opening-lot-cost-trading-currency
type: req
---
An XTB opening lot's cost must be booked in its position's trading currency.

## Why this changed

On 2026-11-01 the author recorded D-159. REQ-ingest-ticker-trading-currency, a default requirement the implementation agent recorded while building D-159 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-159 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. D-23 records an opening lot's cost basis in its original currency. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB position's open price is in its instrument's trading currency, which the export doesn't name. It is the currency of the confirmed ticker's listing (D-116), or, when the listing gives none, the currency of the exchange its market suffix names, and the opening lot's cost is booked in it. A price quoted in a minor unit, such as London's pence, is converted to the major unit.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-159#3

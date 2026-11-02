---
title: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, else of the exchange its market names when that exchange quotes in only one, else the one the owner picked; no currency is assumed, and a price in a minor unit is converted to the major unit.
status: closed
text_ref: docs/DESIGN.md#D-170
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-170
  - lane:strict
semantic_text: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, else of the exchange its market names when that exchange quotes in only one, else the one the owner picked; no currency is assumed, and a price in a minor unit is converted to the major unit.
logic_claims:
  - CLAIM-C75D18E65F168DEA
semantic_clauses:
  - An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, else of the exchange its market names when that exchange quotes in only one, else the one the owner picked; no currency is assumed, and a price in a minor unit is converted to the major unit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c3c34136967471609d9ce966f5659cde388cfec2b439e22e6ecd5c5d297db6c8
semantic_inventory:
  - claim_key: CLAIM-C75D18E65F168DEA
    claim_text: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, else of the exchange its market names when that exchange quotes in only one, else the one the owner picked; no currency is assumed, and a price in a minor unit is converted to the major unit
    role: normative
    status: modeled
    span:
      start: 0
      end: 370
origin:
  kind: agent
  recorded_at: '2026-11-01T23:38:05.926Z'
id: REQ-ingest-ticker-currency-picked
type: req
---
An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, else of the exchange its market names when that exchange quotes in only one, else the one the owner picked; no currency is assumed, and a price in a minor unit is converted to the major unit.

## Context

D-170 corrects D-159: London quotes in pence, pounds and dollars, so its exchange no longer implies pence when the listing names no currency. Recorded by the implementation agent of D-170 (impl/d170) with default firmness: no firm requirement for D-170 existed in kibi when it was built, so this traces the code to the decision directly until the design session imports the firm one.

## Why this changed

The author decided D-170, which corrects D-159 and so replaces REQ-ingest-ticker-trading-currency: that requirement took the exchange's currency whenever the listing named none, pence for London, and is closed.

## Source

> When the listing gives no currency and the exchange its market suffix names quotes in more than one, as London does in pence, pounds and dollars, the owner picks the trading currency when confirming the ticker's ISIN (D-116), and its positions wait until then; no currency is assumed.

docs/DESIGN.md#D-170

> An XTB position's open price is in its instrument's trading currency, which the export doesn't name. It is the currency of the confirmed ticker's listing (D-116), or, when the listing gives none, the currency of the exchange its market suffix names, and the opening lot's cost is booked in it. A price quoted in a minor unit, such as London's pence, is converted to the major unit.

docs/DESIGN.md#D-159

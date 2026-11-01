---
title: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, or of the exchange its market names, with a price in a minor unit converted to the major unit.
status: open
text_ref: docs/DESIGN.md#D-159
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-159
  - lane:strict
semantic_text: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, or of the exchange its market names, with a price in a minor unit converted to the major unit.
logic_claims:
  - CLAIM-71CBF65F08E3B137
semantic_clauses:
  - An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, or of the exchange its market names, with a price in a minor unit converted to the major unit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0d10e3dcaada666ac5376f9096eae8715e65ec36952aefa9ad482ecc022d8a65
semantic_inventory:
  - claim_key: CLAIM-71CBF65F08E3B137
    claim_text: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, or of the exchange its market names, with a price in a minor unit converted to the major unit
    role: normative
    status: modeled
    span:
      start: 0
      end: 273
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:34.894Z'
id: REQ-ingest-ticker-trading-currency
type: req
---
An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, or of the exchange its market names, with a price in a minor unit converted to the major unit.

## Context

D-159 puts an XTB open price in its instrument's trading currency, the listing's or the exchange's. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Source

> An XTB position's open price is in its instrument's trading currency, which the export doesn't name. It is the currency of the confirmed ticker's listing (D-116), or, when the listing gives none, the currency of the exchange its market suffix names, and the opening lot's cost is booked in it. A price quoted in a minor unit, such as London's pence, is converted to the major unit.

docs/DESIGN.md#D-159

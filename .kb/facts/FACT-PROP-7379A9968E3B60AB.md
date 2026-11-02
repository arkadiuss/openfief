---
title: Opening Lot Trading Currency Picked = true
status: active
text_ref: docs/DESIGN.md#D-170
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-170
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_entries
property_key: opening_lot_trading_currency_picked
canonical_key: mcp-kibi-model-requirement:ingest.ticker_entries:opening_lot_trading_currency_picked:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C75D18E65F168DEA
claim_text: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, else of the exchange its market names when that exchange quotes in only one, else the one the owner picked; no currency is assumed, and a price in a minor unit is converted to the major unit.
origin:
  kind: agent
  recorded_at: '2026-11-01T23:38:02.903Z'
id: FACT-PROP-7379A9968E3B60AB
type: fact
---

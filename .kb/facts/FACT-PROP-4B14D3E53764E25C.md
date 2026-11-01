---
title: Opening Lot Trading Currency = true
status: active
text_ref: docs/DESIGN.md#D-159
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-159
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_entries
property_key: opening_lot_trading_currency
canonical_key: mcp-kibi-model-requirement:ingest.ticker_entries:opening_lot_trading_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-71CBF65F08E3B137
claim_text: An opening lot named by ticker with its open price must be booked, once its ticker is confirmed, at its quantity times that price in the trading currency of the confirmed listing, or of the exchange its market names, with a price in a minor unit converted to the major unit.
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:31.419Z'
id: FACT-PROP-4B14D3E53764E25C
type: fact
---

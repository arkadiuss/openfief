---
title: Isin Proposed By Market Data = true
status: active
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_entries
property_key: isin_proposed_by_market_data
canonical_key: mcp-kibi-model-requirement:ingest.ticker_entries:isin_proposed_by_market_data:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-94432C2BAE3F37CE
claim_text: For each ticker reported for the first time, the core must ask the market-data connectors to propose its ISIN, and the eodhd connector must propose the one ISIN EODHD lists for the ticker's code on its market's supported exchange.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:37.526Z'
id: FACT-PROP-3EB285E360BD000A
type: fact
---

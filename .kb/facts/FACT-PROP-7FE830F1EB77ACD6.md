---
title: Writes Prices And Index Levels = true
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.delivery_contract
property_key: writes_prices_and_index_levels
canonical_key: mcp-kibi-model-requirement:ingest.delivery_contract:writes_prices_and_index_levels:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C7564A3EC83ACA2A
claim_text: A connector plugin must write an end-of-day price as its instrument's ISIN and venue symbol, its date and its close with currency, and a market index level as the index's name, its date and its level, as exact decimal strings and YYYY-MM-DD dates.
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:54.766Z'
id: FACT-PROP-7FE830F1EB77ACD6
type: fact
---

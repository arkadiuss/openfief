---
title: Stores Cross Check Prices Apart = true
status: active
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.connector_delivery
property_key: stores_cross_check_prices_apart
canonical_key: mcp-kibi-model-requirement:ingest.connector_delivery:stores_cross_check_prices_apart:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0D1270AC0DDD9ED6
claim_text: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices; a delivery that carries some must fail when no cross-check store is wired rather than drop them.
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:43.185Z'
id: FACT-PROP-9ED1836B2D5CD64B
type: fact
---

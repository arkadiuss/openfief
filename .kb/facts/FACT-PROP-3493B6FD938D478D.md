---
title: Carries Ticker Holdings = true
status: active
text_ref: docs/DESIGN.md#D-155
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-155
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_contract
property_key: carries_ticker_holdings
canonical_key: mcp-kibi-model-requirement:ingest.ticker_contract:carries_ticker_holdings:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2AF0B9208DEBE76D
claim_text: The connector contract must carry the holdings a delivery reports of instruments named by ticker only, and the core must refuse one that names no ticker of its delivery, has no date or has a negative quantity.
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:33.210Z'
id: FACT-PROP-3493B6FD938D478D
type: fact
---

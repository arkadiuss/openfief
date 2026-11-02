---
title: Only To Declared Granted Market Data Connectors = true
status: active
text_ref: design-decisions:D-166#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-166-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.isin_lookup
property_key: only_to_declared_granted_market_data_connectors
canonical_key: mcp-kibi-model-requirement:ingest.isin_lookup:only_to_declared_granted_market_data_connectors:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2E620D9AB6C98199
claim_text: 'A ticker''s ISIN lookup must go only to connectors allowed to learn what the owner holds: those whose manifest declares them market-data connectors and whose grant approves that.'
origin:
  kind: agent
  recorded_at: '2026-11-02T00:55:45.893Z'
id: FACT-PROP-6E59A59E75D4850C
type: fact
---

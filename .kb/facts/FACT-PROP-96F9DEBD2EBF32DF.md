---
title: Asks Declared Market Data Connectors = true
status: active
text_ref: docs/DESIGN.md#D-123
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-123
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.market_data_scope
property_key: asks_declared_market_data_connectors
canonical_key: mcp-kibi-model-requirement:ingest.market_data_scope:asks_declared_market_data_connectors:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-34AFDD3C89B5B844
claim_text: The core must ask every market-data connector instance, one whose manifest declares it and whose grant approves that, to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask every other instance for nothing.
origin:
  kind: agent
  recorded_at: '2026-11-01T10:07:02.451Z'
id: FACT-PROP-96F9DEBD2EBF32DF
type: fact
---

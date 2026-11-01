---
title: Series Sources = true
status: active
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.market_data_service
property_key: series_sources
canonical_key: mcp-kibi-model-requirement:valuation.market_data_service:series_sources:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-76DA85F5A44D246C
claim_text: The MarketDataService must let only a view plugin acting for an owner session list each series' source, change one series' source and set a kind's gap threshold, refusing an invalid series, source or threshold as invalid argument.
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:14.084Z'
id: FACT-PROP-79EA62F65E8FBF30
type: fact
---

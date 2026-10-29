---
title: Owner Ranked Kinds = exchange_rates, index_readings, prices
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-94#1
fact_kind: property_value
subject_key: valuation.market_data_sources
property_key: owner_ranked_kinds
operator: eq
value_type: set
value_set:
  - exchange_rates
  - index_readings
  - prices
canonical_key: docs-design-md:valuation.market_data_sources:owner_ranked_kinds:eq:exchange_rates_index_readings_prices
claim_key: CLAIM-E8238FC7EE4C24F7
claim_text: The owner must rank the sources of each kind of market data, prices, exchange rates and index readings, in priority order.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:48:42.482Z'
id: FACT-PROP-E9B166B02146F3DF
type: fact
---

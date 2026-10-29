---
title: Default First Ranked = exchange_rates=nbp; prices=eodhd; cpi=treasury-bonds; nbp_reference_rate=treasury-bonds
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-4
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-94#4
fact_kind: property_value
subject_key: valuation.market_data_sources
property_key: default_first_ranked
operator: eq
value_type: string
value_string: exchange_rates=nbp; prices=eodhd; cpi=treasury-bonds; nbp_reference_rate=treasury-bonds
canonical_key: docs-design-md:valuation.market_data_sources:default_first_ranked:eq:exchange_rates_nbp_prices_eodhd_cpi_treasury-bonds_nbp_reference_rate_treasury-bonds
claim_key: CLAIM-8D1DCBFD1A770177
claim_text: 'By default the first-party connector for each kind of market data must rank first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.'
origin:
  kind: agent
  recorded_at: '2026-10-29T09:39:29.443Z'
id: FACT-PROP-32BD74E977D25E57
type: fact
---

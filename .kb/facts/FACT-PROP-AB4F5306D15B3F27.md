---
title: Ambiguous Isin Unpriced Until Owner Picks Exchange = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-96-3
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-96#3
fact_kind: property_value
subject_key: plugins.eodhd_connector
property_key: ambiguous_isin_unpriced_until_owner_picks_exchange
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:plugins.eodhd_connector:ambiguous_isin_unpriced_until_owner_picks_exchange:eq:true
claim_key: CLAIM-19E12CFCD60A76F2
claim_text: An ISIN listed on several exchanges the EODHD connector supports must stay unpriced and be reported as ambiguous until the owner picks the exchange.
origin:
  kind: agent
  recorded_at: '2026-10-29T09:43:29.263Z'
id: FACT-PROP-AB4F5306D15B3F27
type: fact
---

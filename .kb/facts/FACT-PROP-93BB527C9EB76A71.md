---
title: Gap Item Per Account = true
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
subject_key: valuation.market_data
property_key: gap_item_per_account
canonical_key: mcp-kibi-model-requirement:valuation.market_data:gap_item_per_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5D6BDA134D2707F2
claim_text: A revaluation must raise a market-data-gap attention item, its subject the account and the series, for each series a holding needs on today whose source's latest value is older than the gap threshold, and clear it for every other series it read.
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:47.393Z'
id: FACT-PROP-93BB527C9EB76A71
type: fact
---

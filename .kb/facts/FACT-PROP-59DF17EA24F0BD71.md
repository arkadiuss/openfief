---
title: Latest Close And Rate = true
status: active
text_ref: docs/DESIGN.md#D-21
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-21
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.market_holding
property_key: latest_close_and_rate
canonical_key: mcp-kibi-model-requirement:valuation.market_holding:latest_close_and_rate:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-396F581CD11DACD3
claim_text: A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to PLN at the latest FX rate on or before the day, rounded half up to grosze.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:42.434Z'
id: FACT-PROP-59DF17EA24F0BD71
type: fact
---

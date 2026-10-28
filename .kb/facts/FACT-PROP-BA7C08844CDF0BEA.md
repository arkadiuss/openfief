---
title: Writes Exact Contract Messages = true
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.delivery
property_key: writes_exact_contract_messages
canonical_key: mcp-kibi-model-requirement:ingest.delivery:writes_exact_contract_messages:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DCDFF191AB58A6CD
claim_text: A connector plugin must write its mapping, FX rates and ingest mode as connector contract messages that keep every amount, rate and quantity exact as a decimal string, every date as YYYY-MM-DD, and must refuse a value with no contract form instead of sending it.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:05.900Z'
id: FACT-PROP-BA7C08844CDF0BEA
type: fact
---

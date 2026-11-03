---
title: Refuses Other Than Pack Currency = true
status: active
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.reporting_currency
property_key: refuses_other_than_pack_currency
canonical_key: mcp-kibi-model-requirement:views.reporting_currency:refuses_other_than_pack_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F66E951806B1CE8C
claim_text: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than the one the jurisdiction pack's stored valuations and rates are in, with a typed unsupported reporting currency error, until cross rates are decided.
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:04.501Z'
id: FACT-PROP-3E561FF11B51286B
type: fact
---

---
title: Served Per Valuation And Total = true
status: active
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.three_values
property_key: served_per_valuation_and_total
canonical_key: mcp-kibi-model-requirement:views.three_values:served_per_valuation_and_total:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7DF79E9014C7A86B
claim_text: The read API must serve each stored valuation with its gross, net and liquidation values or the reason each is missing, totals with the three values left unset when any part lacks one, and the allocation of the value kind asked for, gross by default, listing what lacks it with the reason.
origin:
  kind: agent
  recorded_at: '2026-10-30T00:05:04.264Z'
id: FACT-PROP-BF6ABB0A6AD5B4FC
type: fact
---

---
title: Refuses Rate Without Named Base = true
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
subject_key: valuation.input_validation
property_key: refuses_rate_without_named_base
canonical_key: mcp-kibi-model-requirement:valuation.input_validation:refuses_rate_without_named_base:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-73E2AB4610F596C5
claim_text: Recording a price that is not positive or names no instrument, an FX rate that is not positive, names no base or is quoted in its own currency, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:15.963Z'
id: FACT-PROP-CE052A010E2043D7
type: fact
---

---
title: Parse Bounded = true
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: kernel.decimal
property_key: parse_bounded
canonical_key: mcp-kibi-model-requirement:kernel.decimal:parse_bounded:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-374FFF6BEF80FC12
claim_text: ParseDecimal must refuse text longer than 256 bytes or with a scientific exponent beyond plus or minus 100, before building the number, and its errors must never quote the text.
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:43.522Z'
id: FACT-PROP-BC830C661CAE9BC4
type: fact
---

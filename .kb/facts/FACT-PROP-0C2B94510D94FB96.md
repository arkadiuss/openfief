---
title: Refuses Hostile Statements = true
status: active
text_ref: docs/DESIGN.md#3.5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-3-5
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ibkr_flex
property_key: refuses_hostile_statements
canonical_key: mcp-kibi-model-requirement:ingest.ibkr_flex:refuses_hostile_statements:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EDF43E2B11A65C0F
claim_text: The IBKR Flex parser must refuse a statement over its size, nesting depth, element count, attribute count or attribute length limits, any DOCTYPE, DTD or entity, and non-UTF-8 input, and must read amounts only as exact decimals in plain notation.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:24:06.079Z'
id: FACT-PROP-0C2B94510D94FB96
type: fact
---

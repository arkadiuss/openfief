---
title: Allow Listed Attributes Only = true
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: telemetry.traces
property_key: allow_listed_attributes_only
canonical_key: mcp-kibi-model-requirement:telemetry.traces:allow_listed_attributes_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1706FF3D39230191
claim_text: 'Exported spans must never carry money amounts, quantities, account numbers, ISINs, tokens, session IDs, secrets, SQL statements or their arguments, or error messages: only an allow-list of attributes may leave the process.'
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:12.918Z'
id: FACT-PROP-3E9F6BA9A9650237
type: fact
---
